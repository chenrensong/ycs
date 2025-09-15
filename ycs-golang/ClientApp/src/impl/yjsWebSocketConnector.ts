import * as Y from 'yjs';
import * as awarenessProtocol from 'y-protocols/awareness.js';
import { byteArrayToString, stringToByteArray } from '../util/encodingUtils';

interface YjsMessage {
  readonly clock: number;
  readonly data: string;
  readonly inReplyTo?: string;
}

interface YjsPendingReceivedMessage {
  readonly clock: number;
  readonly data: string;
  readonly type: YjsMessageType;
  readonly inReplyTo?: string;
}

enum YjsMessageType {
  GetMissing = 'GetMissing',        // SyncStep1
  Update = 'Update',                // SyncStep2
  QueryAwareness = 'QueryAwareness',    // Other clients will broadcast their awareness info.
  UpdateAwareness = 'UpdateAwareness'   // Broadcast awareness info.
}

export class YjsWebSocketConnector {
  private _doc: Y.Doc;
  private _ws: WebSocket | null = null;
  private _awareness: awarenessProtocol.Awareness;
  private _url: string;

  private _synced: boolean = false;
  private _connected: boolean = false;

  private _receiveQueue: YjsPendingReceivedMessage[] = [];
  private _clientClock: number = -1;
  private _serverClock: number = -1;
  private _resyncInterval?: number = undefined;
  private _reconnectTimeout?: number = undefined;

  constructor(doc: Y.Doc, url: string) {
    this._doc = doc;
    this._url = url;

    this._awareness = new awarenessProtocol.Awareness(this._doc);
    this._awareness.on('update', this._awarenessUpdateHandler);

    this._doc.on('updateV2', this._yDocUpdateV2Handler);

    this._connect();

    // Resync every 10 seconds.
    this._resyncInterval = window.setInterval(async () => {
      if (this._connected) {
        await this._requestMissingAsync();
        await this._requestAndBroadcastAwareness();
      }
    }, 10000);
  }

  public get awareness(): awarenessProtocol.Awareness {
    return this._awareness;
  }

  private get connected(): boolean {
    return this._connected && this._ws !== null && this._ws.readyState === WebSocket.OPEN;
  }

  private _connect(): void {
    try {
      this._ws = new WebSocket(this._url);
      
      this._ws.onopen = () => {
        console.log('WebSocket connected');
        this._connected = true;
        this._resetConnectionAsync();
      };

      this._ws.onclose = () => {
        console.log('WebSocket disconnected');
        this._connected = false;
        this._attemptReconnect();
      };

      this._ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this._connected = false;
      };

      this._ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          this._handleMessage(message);
        } catch (err) {
          console.error('Error parsing WebSocket message:', err);
        }
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this._attemptReconnect();
    }
  }

  private _attemptReconnect(): void {
    if (this._reconnectTimeout) {
      clearTimeout(this._reconnectTimeout);
    }

    this._reconnectTimeout = window.setTimeout(() => {
      console.log('Attempting to reconnect...');
      this._connect();
    }, 3000);
  }

  private _handleMessage(message: any): void {
    if (message.type === YjsMessageType.GetMissing && message.data) {
      this._onGetMissingReceived(message.data);
    } else if (message.type === YjsMessageType.Update && message.data) {
      this._onUpdateReceived(message.data);
    } else if (message.type === YjsMessageType.QueryAwareness) {
      this._onQueryAwareness();
    } else if (message.type === YjsMessageType.UpdateAwareness && message.data) {
      this._onUpdateAwareness(message.data);
    }
  }

  public destroy(): void {
    this._doc.off('updateV2', this._yDocUpdateV2Handler);
    this._awareness.off('update', this._awarenessUpdateHandler);

    awarenessProtocol.removeAwarenessStates(this._awareness, [this._doc.clientID], this);

    if (this._resyncInterval) {
      window.clearInterval(this._resyncInterval);
      this._resyncInterval = undefined;
    }

    if (this._reconnectTimeout) {
      window.clearTimeout(this._reconnectTimeout);
      this._reconnectTimeout = undefined;
    }

    if (this._ws) {
      this._ws.close();
      this._ws = null;
    }

    this._connected = false;
  }

  private _yDocUpdateV2Handler = (updateMessage: Uint8Array, origin: object | undefined): void => {
    if (origin !== this && origin !== 'websocket') {
      this._sendMessageAsync(YjsMessageType.Update, updateMessage, undefined);
    }
  };

  private _awarenessUpdateHandler = ({ added, updated, removed }): void => {
    const changedClients = added.concat(updated).concat(removed);
    const update = awarenessProtocol.encodeAwarenessUpdate(this._awareness, changedClients);
    this._sendMessageAsync(YjsMessageType.UpdateAwareness, update, undefined);
  };

  // SyncStep1
  private _onGetMissingReceived = (data: YjsMessage): void => {
    if (data) {
      this._enqueueAndProcessMessages(data, YjsMessageType.GetMissing);
    }
  };

  // SyncStep2
  private _onUpdateReceived = (data: YjsMessage): void => {
    if (data) {
      this._enqueueAndProcessMessages(data, YjsMessageType.Update);
    }
  };

  private _onQueryAwareness = (): void => {
    const update = awarenessProtocol.encodeAwarenessUpdate(this._awareness, Array.from(this._awareness.getStates().keys()));
    this._sendMessageAsync(YjsMessageType.UpdateAwareness, update, undefined);
  };

  private _onUpdateAwareness = (data: string): void => {
    if (data) {
      const update = stringToByteArray(data);
      awarenessProtocol.applyAwarenessUpdate(this._awareness, update, this);
    }
  };

  private async _resetConnectionAsync(): Promise<void> {
    this._synced = false;
    this._clientClock = -1;
    this._serverClock = -1;
    this._receiveQueue = [];

    if (this.connected) {
      await this._requestMissingAsync();
      await this._requestAndBroadcastAwareness();
    } else {
      // Update awareness (all users except local left).
      awarenessProtocol.removeAwarenessStates(
        this._awareness,
        Array.from(this._awareness.getStates().keys()).filter(client => client !== this._doc.clientID),
        this);
    }
  }

  private async _requestMissingAsync(): Promise<void> {
    if (!this.connected) {
      return;
    }

    const stateVector = Y.encodeStateVectorV2(this._doc);
    await this._sendMessageAsync(YjsMessageType.GetMissing, stateVector, undefined);
  }

  private async _requestAndBroadcastAwareness(): Promise<void> {
    if (!this.connected) {
      return;
    }

    await this._sendMessageAsync(YjsMessageType.QueryAwareness, undefined, undefined);

    if (this._awareness.getLocalState() !== null) {
      const update = awarenessProtocol.encodeAwarenessUpdate(this._awareness, [this._doc.clientID]);
      await this._sendMessageAsync(YjsMessageType.UpdateAwareness, update, undefined);
    }
  }

  private async _sendMessageAsync(
    type: YjsMessageType,
    data: Uint8Array | undefined,
    inReplyTo: YjsMessageType | undefined
  ): Promise<void> {
    if (!this.connected || !this._ws) {
      return;
    }

    if (data === undefined) {
      data = new Uint8Array();
    }

    try {
      switch (type) {
        case YjsMessageType.GetMissing:
        case YjsMessageType.Update:
          const message = {
            type: type,
            data: {
              clock: ++this._clientClock,
              data: byteArrayToString(data),
              inReplyTo: inReplyTo
            }
          };
          this._ws.send(JSON.stringify(message));
          break;
        case YjsMessageType.QueryAwareness:
        case YjsMessageType.UpdateAwareness:
          // Awareness has its own internal clock, no need to duplicate it in the message.
          const awarenessMessage = {
            type: type,
            data: byteArrayToString(data)
          };
          this._ws.send(JSON.stringify(awarenessMessage));
          break;
        default:
          throw new Error(`Unknown message type: ${type}`);
      }
    } catch (error) {
      console.error('Failed to send WebSocket message:', error);
    }
  }

  private _enqueueAndProcessMessages(data: YjsMessage, type: YjsMessageType): void {
    // Invalid data.
    if (!data || data.clock === undefined || !data.data) {
      return;
    }

    this._receiveQueue.push({
      clock: data.clock,
      data: data.data,
      inReplyTo: data.inReplyTo,
      type
    });

    // Remove messages that we should've processed already.
    this._receiveQueue = this._receiveQueue.filter(msg => msg.clock > this._serverClock);

    // Sort queue by sequence number.
    this._receiveQueue.sort((a, b) => a.clock - b.clock);

    // We can fast-forward server clock if we're 'stuck' and/or have pending
    // UpdateV2 (SyncStep2) messages - they indicate the reply on the initial/periodic
    // sync that will eventually make previous updates no-op.
    const isInitialSyncMessage = (msg: YjsPendingReceivedMessage) =>
      msg.type === YjsMessageType.Update && msg.inReplyTo === YjsMessageType.GetMissing;
    if (this._receiveQueue.some(isInitialSyncMessage)) {
      while (this._receiveQueue.length > 0 && !isInitialSyncMessage(this._receiveQueue[0])) {
        this._receiveQueue.shift();
      }
      if (this._receiveQueue.length > 0) {
        this._serverClock = this._receiveQueue[0].clock - 1;
      }
    }

    this._doc.transact(() => {
      while (this._receiveQueue.length > 0) {
        const msg = this._receiveQueue[0];

        // Check for potential duplicates (unlikely to happen).
        if (msg.clock === this._serverClock) {
          this._receiveQueue.shift();
          continue;
        }

        // Check whether the next message is something we can apply now.
        if (msg.clock !== this._serverClock + 1) {
          break;
        }

        try {
          switch (msg.type) {
            // SyncStep1
            case YjsMessageType.GetMissing:
              // Reply with SyncStep2 on SyncStep1.
              const targetStateVector = stringToByteArray(msg.data);
              const update = Y.encodeStateAsUpdateV2(this._doc, targetStateVector);
              this._sendMessageAsync(YjsMessageType.Update, update, YjsMessageType.GetMissing);
              break;
            // SyncStep2
            case YjsMessageType.Update:
              // Skip all updates received until the missing blocks are applied.
              if (this._synced || msg.inReplyTo === YjsMessageType.GetMissing) {
                const update = stringToByteArray(msg.data);
                Y.applyUpdateV2(this._doc, update, 'websocket');

                if (msg.inReplyTo === YjsMessageType.GetMissing) {
                  this._synced = true;
                }
              }
              break;
            default:
              throw new Error(`Unsupported Yjs message type: ${msg.type}`);
          }
        } catch (e) {
          console.error(e);
          throw e;
        } finally {
          // Remove the message from the queue.
          this._receiveQueue.shift();
          this._serverClock++;
        }
      }
    }, this);
  }
}
