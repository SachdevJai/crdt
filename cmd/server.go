package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"crdt/pkg/rga" // Import the RGA package

	"github.com/coder/websocket"
	"github.com/joho/godotenv"
)

// RGAMessage represents a message for the RGA operations.
type RGAMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Position  int    `json:"position"`
	Value     string `json:"value"`
}

var (
	documentFile = "document.json" // Path to the file where document is stored
)

// LoadDocument loads the document from the file or returns a new one if it doesn't exist.
func loadDocument() (*rga.RGA, error) {
	// Check if the document file exists
	if _, err := os.Stat(documentFile); err == nil {
		// File exists, load the document
		data, err := os.ReadFile(documentFile)
		if err != nil {
			return nil, err
		}

		var document []string
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, err
		}

		// Rebuild the RGA from the saved document
		rgaDoc := rga.NewRGA()
		for _, char := range document {
			rgaDoc.Insert(len(rgaDoc.GetDocument()), char)
		}
		return rgaDoc, nil
	}

	// File does not exist, return a new RGA document
	return rga.NewRGA(), nil
}

// SaveDocument saves the current document to a file.
func saveDocument(rgaDoc *rga.RGA) error {
	document := rgaDoc.GetDocument()
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}

	// Save the document to the file
	if err := os.WriteFile(documentFile, data, 0644); err != nil {
		return err
	}
	return nil
}

// wsHandler handles WebSocket connections from clients.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "closing")

	log.Println("Client connected")

	// Load the RGA document from the file
	rgaDoc, err := loadDocument()
	if err != nil {
		log.Println("Error loading document:", err)
		return
	}

	// Send the current document to the client upon connection
	document := rgaDoc.GetDocument()
	documentJSON, _ := json.Marshal(document)
	if err := conn.Write(r.Context(), websocket.MessageText, documentJSON); err != nil {
		log.Println("Error sending initial document to client:", err)
		return
	}

	// Handle incoming messages
	for {
		_, data, err := conn.Read(r.Context())
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		log.Printf("Received: %s", string(data))

		var msg RGAMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		// Apply the operation to the RGA
		if msg.Type == "insert" {
			rgaDoc.Insert(msg.Position, msg.Value)
		} else if msg.Type == "delete" {
			rgaDoc.Delete(msg.Position)
		}

		// Save the updated document to the file
		if err := saveDocument(rgaDoc); err != nil {
			log.Println("Error saving document:", err)
		}

		// Send the updated document back to the client
		document = rgaDoc.GetDocument()
		documentJSON, _ := json.Marshal(document)
		if err := conn.Write(r.Context(), websocket.MessageText, documentJSON); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	addr := fmt.Sprintf(":%s", os.Getenv("PORT"))
	http.HandleFunc("/ws", wsHandler)

	fmt.Printf("WebSocket server running at ws://localhost%s/ws\n", addr)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
