package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hypebeast/go-osc/osc"
)

// OSCMessage represents an OSC message from the web frontend
type OSCMessage struct {
	Address string        `json:"address"`
	Args    []interface{} `json:"args"`
}

func main() {
	// Create OSC client for SuperCollider sclang (port 57120)
	client := osc.NewClient("localhost", 57120)

	// HTTP endpoint for receiving OSC messages from browser
	http.HandleFunc("/osc", func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for local development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var msg OSCMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Create OSC message
		oscMsg := osc.NewMessage(msg.Address)
		for _, arg := range msg.Args {
			oscMsg.Append(arg)
		}

		// Send to SuperCollider
		if err := client.Send(oscMsg); err != nil {
			log.Printf("Error sending OSC: %v", err)
			http.Error(w, "Failed to send OSC", http.StatusInternalServerError)
			return
		}

		log.Printf("Sent OSC: %s %v", msg.Address, msg.Args)
		w.WriteHeader(http.StatusOK)
	})

	log.Println("OSC Bridge running on :8080")
	log.Println("Forwarding HTTP POST → OSC UDP to localhost:57120")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
