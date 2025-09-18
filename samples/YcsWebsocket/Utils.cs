using System.Collections.Concurrent;
using System.Net.WebSockets;
using Ycs.Contracts;

namespace YcsWebsocket
{
    // 定义常量
    public static class Constants
    {
        public const int MessageSync = 0;
        public const int MessageAwareness = 1;
        public const int PingTimeout = 30000;
        public const int CallbackDebounceWait = 2000;
        public const int CallbackDebounceMaxWait = 10000;
    }

    // 共享文档类
    public class WSSharedDoc : Ycs.Core.YDoc
    {
        public string Name { get; }
        public Dictionary<WebSocket, HashSet<int>> Conns { get; } = new Dictionary<WebSocket, HashSet<int>>();
        public Dictionary<int, object> AwarenessStates { get; } = new Dictionary<int, object>();
        public bool GCEnabled { get; set; } = true;

        public WSSharedDoc(string name)
        {
            Name = name;
        }

        internal object GetXmlFragment(string objName)
        {
            return "";
        }

        internal object GetXmlElement(string objName)
        {
            return "";
        }
    }

    // 主服务类
    public class YjsWebSocketServer
    {
        private static readonly ConcurrentDictionary<string, WSSharedDoc> Docs = new ConcurrentDictionary<string, WSSharedDoc>();
        private static IPersistence _persistence;
        private static Func<WSSharedDoc, Task> _contentInitializer = doc => Task.CompletedTask;

        public static void SetPersistence(IPersistence persistence)
        {
            _persistence = persistence;
        }

        public static IPersistence GetPersistence() => _persistence;

        public static void SetContentInitializer(Func<WSSharedDoc, Task> initializer)
        {
            _contentInitializer = initializer;
        }

        public static WSSharedDoc GetYDoc(string docName, bool gc = true)
        {
            return Docs.GetOrAdd(docName, name =>
            {
                var doc = new WSSharedDoc(name) { GCEnabled = gc };
                _persistence?.BindState(name, doc);
                _contentInitializer(doc).Wait();
                return doc;
            });
        }

        public static async Task SetupWSConnection(WebSocket webSocket, HttpContext context, string docName = null, bool gc = true)
        {
            docName ??= context.Request.Path.Value?.TrimStart('/').Split('?')[0] ?? Guid.NewGuid().ToString();

            var doc = GetYDoc(docName, gc);
            doc.Conns[webSocket] = new HashSet<int>();

            // 处理消息
            var buffer = new byte[1024 * 4];
            var pongReceived = true;
            var pingTimer = new Timer(async _ =>
            {
                if (!pongReceived)
                {
                    await CloseConnection(doc, webSocket);
                    return;
                }

                if (doc.Conns.ContainsKey(webSocket))
                {
                    pongReceived = false;
                    try
                    {
                        await webSocket.SendAsync(new ArraySegment<byte>(Array.Empty<byte>()),
                            WebSocketMessageType.Binary, true, CancellationToken.None);
                    }
                    catch
                    {
                        await CloseConnection(doc, webSocket);
                    }
                }
            }, null, Constants.PingTimeout, Constants.PingTimeout);

            try
            {
                // 发送初始同步消息
                await SendSyncStep1(doc, webSocket);
                if (doc.AwarenessStates.Count > 0)
                {
                    await SendAwarenessUpdate(doc, webSocket, doc.AwarenessStates.Keys);
                }

                while (webSocket.State == WebSocketState.Open)
                {
                    var result = await webSocket.ReceiveAsync(new ArraySegment<byte>(buffer), CancellationToken.None);
                    if (result.MessageType == WebSocketMessageType.Close)
                    {
                        await CloseConnection(doc, webSocket);
                        break;
                    }

                    await ProcessMessage(doc, webSocket, buffer, result.Count);
                }
            }
            finally
            {
                await CloseConnection(doc, webSocket);
                pingTimer.Dispose();
            }
        }

        private static async Task ProcessMessage(WSSharedDoc doc, WebSocket webSocket, byte[] buffer, int count)
        {
            try
            {
                var message = new ArraySegment<byte>(buffer, 0, count);
                using var memoryStream = new System.IO.MemoryStream(message.ToArray());
                using var reader = new System.IO.BinaryReader(memoryStream);

                var messageType = ReadVarUint(reader);
                switch (messageType)
                {
                    case Constants.MessageSync:
                        await ProcessSyncMessage(doc, webSocket, reader);
                        break;
                    case Constants.MessageAwareness:
                        ProcessAwarenessMessage(doc, webSocket, reader);
                        break;
                }
            }
            catch (Exception ex)
            {
                Console.Error.WriteLine($"Error processing message: {ex.Message}");
                await CloseConnection(doc, webSocket);
            }
        }

        private static async Task ProcessSyncMessage(WSSharedDoc doc, WebSocket webSocket, System.IO.BinaryReader reader)
        {
            // 这里需要实现Yjs同步协议的具体逻辑
            // 简化为直接转发消息给其他客户端
            var response = new System.IO.MemoryStream();
            WriteVarUint(response, Constants.MessageSync);
            response.Write(reader.ReadBytes((int)(reader.BaseStream.Length - reader.BaseStream.Position)));

            await BroadcastToOthers(doc, webSocket, response.ToArray());
        }

        private static void ProcessAwarenessMessage(WSSharedDoc doc, WebSocket webSocket, System.IO.BinaryReader reader)
        {
            // 处理awareness更新
            var update = ReadVarUint8Array(reader);
            // 简化为直接更新awareness状态
            var clientId = BitConverter.ToInt32(update, 0);
            doc.AwarenessStates[clientId] = new { /* 状态数据 */ };
            doc.Conns[webSocket].Add(clientId);

            // 广播给其他客户端
            _ = BroadcastAwarenessUpdate(doc, webSocket, new[] { clientId });
        }

        private static async Task SendSyncStep1(WSSharedDoc doc, WebSocket webSocket)
        {
            var message = new System.IO.MemoryStream();
            WriteVarUint(message, Constants.MessageSync);
            // 这里需要添加实际的同步数据
            await Send(doc, webSocket, message.ToArray());
        }

        private static async Task SendAwarenessUpdate(WSSharedDoc doc, WebSocket webSocket, IEnumerable<int> clientIds)
        {
            var message = new System.IO.MemoryStream();
            WriteVarUint(message, Constants.MessageAwareness);
            // 这里需要添加实际的awareness数据
            await Send(doc, webSocket, message.ToArray());
        }

        private static async Task BroadcastToOthers(WSSharedDoc doc, WebSocket sender, byte[] message)
        {
            foreach (var conn in doc.Conns.Keys)
            {
                if (conn != sender && conn.State == WebSocketState.Open)
                {
                    await Send(doc, conn, message);
                }
            }
        }

        private static async Task BroadcastAwarenessUpdate(WSSharedDoc doc, WebSocket sender, IEnumerable<int> clientIds)
        {
            foreach (var conn in doc.Conns.Keys)
            {
                if (conn != sender && conn.State == WebSocketState.Open)
                {
                    await SendAwarenessUpdate(doc, conn, clientIds);
                }
            }
        }

        private static async Task Send(WSSharedDoc doc, WebSocket webSocket, byte[] message)
        {
            if (webSocket.State != WebSocketState.Open)
            {
                await CloseConnection(doc, webSocket);
                return;
            }

            try
            {
                await webSocket.SendAsync(new ArraySegment<byte>(message),
                    WebSocketMessageType.Binary, true, CancellationToken.None);
            }
            catch
            {
                await CloseConnection(doc, webSocket);
            }
        }

        private static async Task CloseConnection(WSSharedDoc doc, WebSocket webSocket)
        {
            if (doc.Conns.TryGetValue(webSocket, out var controlledIds))
            {
                doc.Conns.Remove(webSocket);
                foreach (var clientId in controlledIds)
                {
                    doc.AwarenessStates.Remove(clientId);
                }

                if (doc.Conns.Count == 0 && _persistence != null)
                {
                    await _persistence.WriteState(doc.Name, doc);
                    Docs.TryRemove(doc.Name, out _);
                }
            }

            try
            {
                await webSocket.CloseAsync(WebSocketCloseStatus.NormalClosure, "Closing", CancellationToken.None);
            }
            catch
            {
                // 忽略关闭错误
            }
        }

        // 以下是Yjs协议相关的辅助方法
        private static uint ReadVarUint(System.IO.BinaryReader reader)
        {
            // 实现Yjs的变长整数读取
            uint result = 0;
            byte current;
            int shift = 0;
            do
            {
                current = reader.ReadByte();
                result |= (uint)(current & 0x7f) << shift;
                shift += 7;
            } while ((current & 0x80) != 0);
            return result;
        }

        private static void WriteVarUint(System.IO.Stream stream, uint value)
        {
            // 实现Yjs的变长整数写入
            do
            {
                byte current = (byte)(value & 0x7f);
                value >>= 7;
                if (value != 0) current |= 0x80;
                stream.WriteByte(current);
            } while (value != 0);
        }

        private static byte[] ReadVarUint8Array(System.IO.BinaryReader reader)
        {
            var length = (int)ReadVarUint(reader);
            return reader.ReadBytes(length);
        }
    }

    // 持久化接口
    public interface IPersistence
    {
        void BindState(string docName, WSSharedDoc doc);
        Task WriteState(string docName, WSSharedDoc doc);
    }

    // ASP.NET Core 集成
    public static class YjsWebSocketMiddleware
    {
        public static IApplicationBuilder UseYjsWebSocketServer(this IApplicationBuilder app)
        {
            app.UseWebSockets();
            app.Use(async (context, next) =>
            {
                if (context.WebSockets.IsWebSocketRequest)
                {
                    var webSocket = await context.WebSockets.AcceptWebSocketAsync();
                    await YjsWebSocketServer.SetupWSConnection(webSocket, context);
                }
                else
                {
                    await next();
                }
            });
            return app;
        }
    }
}
