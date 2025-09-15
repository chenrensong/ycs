package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"ycs/contracts"
	"ycs/core"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// YjsCommandType represents the command types for Yjs messages
type YjsCommandType string

const (
	GetMissing YjsCommandType = "GetMissing"
	Update     YjsCommandType = "Update"
)

// YjsMessage represents a message structure for Yjs communication
type YjsMessage struct {
	Clock     int64           `json:"clock"`
	Data      string          `json:"data"`
	InReplyTo *YjsCommandType `json:"inReplyTo,omitempty"`
}

// MessageToProcess represents a message to be processed
type MessageToProcess struct {
	Command   YjsCommandType
	InReplyTo *YjsCommandType
	Data      string
}

// ClientContext manages the state for each connected client
type ClientContext struct {
	synced      bool
	serverClock int64
	clientClock int64
	messages    map[int64]*MessageToProcess
	conn        *websocket.Conn
	mutex       sync.RWMutex
}

func NewClientContext(conn *websocket.Conn) *ClientContext {
	return &ClientContext{
		synced:      false,
		serverClock: -1,
		clientClock: -1,
		messages:    make(map[int64]*MessageToProcess),
		conn:        conn,
	}
}

func (cc *ClientContext) IncrementAndGetServerClock() int64 {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.serverClock++
	return cc.serverClock
}

func (cc *ClientContext) IncrementAndGetClientClock() int64 {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.clientClock++
	return cc.clientClock
}

func (cc *ClientContext) ReassignClientClock(clock int64) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.clientClock = clock
}

func (cc *ClientContext) IsSynced() bool {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()
	return cc.synced
}

func (cc *ClientContext) SetSynced(synced bool) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.synced = synced
}

// YcsManager manages the YCS document and client connections
type YcsManager struct {
	doc     *core.YDoc
	clients map[string]*ClientContext
	mutex   sync.RWMutex
}

func NewYcsManager() *YcsManager {
	manager := &YcsManager{
		doc:     core.NewYDoc(contracts.YDocOptions{}),
		clients: make(map[string]*ClientContext),
	}

	// Prepopulate document with data (like C# version)
	text := manager.doc.GetText("monaco")
	text.Insert(0, "Hello, world!")

	// Set up update handler
	manager.doc.OnUpdateV2(func(update []byte, origin interface{}, transaction contracts.ITransaction) {
		if update == nil || len(update) == 0 {
			return
		}

		encodedUpdate := base64.StdEncoding.EncodeToString(update)

		// Send update to all synced clients
		manager.mutex.RLock()
		clients := make([]*ClientContext, 0, len(manager.clients))
		for _, client := range manager.clients {
			if client.IsSynced() {
				clients = append(clients, client)
			}
		}
		manager.mutex.RUnlock()

		for _, client := range clients {
			msg := YjsMessage{
				Clock: client.IncrementAndGetServerClock(),
				Data:  encodedUpdate,
			}

			response := map[string]interface{}{
				"type": string(Update),
				"data": msg,
			}

			go func(c *ClientContext) {
				if err := c.conn.WriteJSON(response); err != nil {
					log.Printf("Error sending update to client: %v", err)
				}
			}(client)
		}
	})

	return manager
}

func (ym *YcsManager) HandleClientConnected(clientID string, conn *websocket.Conn) {
	ym.mutex.Lock()
	defer ym.mutex.Unlock()
	ym.clients[clientID] = NewClientContext(conn)
	log.Printf("Client connected: %s", clientID)
}

func (ym *YcsManager) HandleClientDisconnected(clientID string) {
	ym.mutex.Lock()
	defer ym.mutex.Unlock()
	delete(ym.clients, clientID)
	log.Printf("Client disconnected: %s", clientID)
}

func (ym *YcsManager) ProcessMessage(clientID string, clock int64, message *MessageToProcess) error {
	ym.mutex.RLock()
	client, exists := ym.clients[clientID]
	ym.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	client.mutex.Lock()
	client.messages[clock] = message
	client.mutex.Unlock()

	// Process messages in order
	return ym.processMessagesInOrder(client)
}

func (ym *YcsManager) processMessagesInOrder(client *ClientContext) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	// Process messages in chronological order
	for {
		nextClock := client.clientClock + 1
		message, exists := client.messages[nextClock]
		if !exists {
			break
		}

		switch message.Command {
		case GetMissing:
			if err := ym.handleGetMissing(client, message); err != nil {
				return err
			}
		case Update:
			if err := ym.handleUpdate(client, message); err != nil {
				return err
			}
		}

		client.clientClock++
		delete(client.messages, nextClock)
	}

	return nil
}

func (ym *YcsManager) handleGetMissing(client *ClientContext, message *MessageToProcess) error {
	// Decode state vector
	decodedStateVector, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return err
	}

	// Generate update and state vector
	update := ym.doc.EncodeStateAsUpdateV2(decodedStateVector)
	stateVector := ym.doc.EncodeStateVectorV2()

	// Send SyncStep2 (Update) message
	getMissingType := GetMissing
	syncStep2Message := YjsMessage{
		Clock:     client.IncrementAndGetServerClock(),
		Data:      base64.StdEncoding.EncodeToString(update),
		InReplyTo: &getMissingType,
	}

	response2 := map[string]interface{}{
		"type": string(Update),
		"data": syncStep2Message,
	}

	if err := client.conn.WriteJSON(response2); err != nil {
		return err
	}

	// Send SyncStep1 (GetMissing) message
	syncStep1Message := YjsMessage{
		Clock:     client.IncrementAndGetServerClock(),
		Data:      base64.StdEncoding.EncodeToString(stateVector),
		InReplyTo: &getMissingType,
	}

	response1 := map[string]interface{}{
		"type": string(GetMissing),
		"data": syncStep1Message,
	}

	return client.conn.WriteJSON(response1)
}

func (ym *YcsManager) handleUpdate(client *ClientContext, message *MessageToProcess) error {
	// Only process updates if client is synced or this is a sync response
	getMissingType := GetMissing
	if !client.IsSynced() && (message.InReplyTo == nil || *message.InReplyTo != getMissingType) {
		return nil
	}

	// Decode update
	update, err := base64.StdEncoding.DecodeString(message.Data)
	if err != nil {
		return err
	}

	// Apply update to document
	ym.doc.ApplyUpdateV2(update, "websocket", false)

	// Mark client as synced if this was a sync response
	if message.InReplyTo != nil && *message.InReplyTo == GetMissing {
		client.SetSynced(true)
	}

	return nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

var ycsManager = NewYcsManager()

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	ycsManager.HandleClientConnected(clientID, conn)
	defer ycsManager.HandleClientDisconnected(clientID)

	for {
		var rawMessage map[string]interface{}
		if err := conn.ReadJSON(&rawMessage); err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		messageType, ok := rawMessage["type"].(string)
		if !ok {
			log.Printf("Invalid message format: missing type")
			continue
		}

		dataRaw, ok := rawMessage["data"]
		if !ok {
			log.Printf("Invalid message format: missing data")
			continue
		}

		// Handle both string and object data formats
		var yjsMessage YjsMessage
		switch v := dataRaw.(type) {
		case string:
			if err := json.Unmarshal([]byte(v), &yjsMessage); err != nil {
				log.Printf("Error parsing YJS message from string: %v", err)
				continue
			}
		case map[string]interface{}:
			// Convert map to YjsMessage
			if clock, ok := v["clock"].(float64); ok {
				yjsMessage.Clock = int64(clock)
			}
			if data, ok := v["data"].(string); ok {
				yjsMessage.Data = data
			}
			if inReplyTo, ok := v["inReplyTo"].(string); ok && inReplyTo != "" {
				replyType := YjsCommandType(inReplyTo)
				yjsMessage.InReplyTo = &replyType
			}
		default:
			log.Printf("Invalid data format: %T", dataRaw)
			continue
		}

		command := YjsCommandType(messageType)
		messageToProcess := &MessageToProcess{
			Command:   command,
			InReplyTo: yjsMessage.InReplyTo,
			Data:      yjsMessage.Data,
		}

		if err := ycsManager.ProcessMessage(clientID, yjsMessage.Clock, messageToProcess); err != nil {
			log.Printf("Error processing message: %v", err)
		}
	}
}

func handleSPA(w http.ResponseWriter, r *http.Request) {
	// Check if the ClientApp/build directory exists and serve from there
	// Otherwise, serve a simple fallback page
	buildPath := "./ClientApp/build"
	indexPath := buildPath + "/index.html"

	// Check if built React app exists
	if _, err := http.Dir(buildPath).Open("index.html"); err == nil {
		// Serve the React app
		http.ServeFile(w, r, indexPath)
	} else {
		// Fallback to simple HTML page
		handleSimpleIndex(w, r)
	}
}

func handleSimpleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>YCS Golang - Please Build React App</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .warning { background: #fff3cd; border: 1px solid #ffeaa7; padding: 20px; border-radius: 5px; }
        .instructions { background: #f8f9fa; padding: 20px; border-radius: 5px; margin-top: 20px; }
        code { background: #e9ecef; padding: 2px 5px; border-radius: 3px; }
        pre { background: #e9ecef; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="container">
        <h1>YCS Golang Server</h1>
        
        <div class="warning">
            <h3>⚠️ React App Not Built</h3>
            <p>The React ClientApp needs to be built before it can be served. Please follow the instructions below.</p>
        </div>

        <div class="instructions">
            <h3>🚀 How to Build and Run the React App:</h3>
            <ol>
                <li>Open a new terminal and navigate to the ClientApp directory:
                    <pre>cd ClientApp</pre>
                </li>
                <li>Install dependencies:
                    <pre>npm install</pre>
                </li>
                <li>Build the React app:
                    <pre>npm run build</pre>
                </li>
                <li>Refresh this page to see the React app!</li>
            </ol>
            
            <h4>For Development:</h4>
            <p>You can also run the React app in development mode:</p>
            <pre>cd ClientApp
npm start</pre>
            <p>This will start the React dev server on <code>http://localhost:3000</code></p>
        </div>

        <div class="instructions">
            <h3>📡 WebSocket Connection</h3>
            <p>The Golang server is running on <strong>http://localhost:8080</strong></p>
            <p>WebSocket endpoint: <strong>ws://localhost:8080/ws</strong></p>
            <p>The React app will automatically connect to this WebSocket endpoint for real-time collaboration.</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func main() {
	// Initialize the system
	core.Initialize()

	log.Printf("Starting YCS Golang server...")
	log.Printf("Created YDoc with client ID: %d", ycsManager.doc.GetClientID())

	// Setup routes
	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", handleWebSocket)

	// Serve React app static files if they exist
	buildPath := "./ClientApp/build"
	if _, err := http.Dir(buildPath).Open("index.html"); err == nil {
		// Serve static files from build directory
		r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(buildPath+"/static/"))))
		// Serve other assets
		r.PathPrefix("/").Handler(http.FileServer(http.Dir(buildPath))).Methods("GET")
	}

	// Fallback route for SPA (must be last)
	r.PathPrefix("/").HandlerFunc(handleSPA).Methods("GET")

	// Additional static file serving
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./static/"))))

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Printf("Open http://localhost%s in your browser to test", port)

	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
