package main

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
)

func main() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)

	if err != nil {
		log.Printf("Websock connect failed: %v", err)
		return
	}

	defer conn.CloseNow()

	for {
		messageType, data, err := conn.Read(r.Context())

		if err != nil {
			return
		}

		if err := conn.Write(r.Context(), messageType, data); err != nil {
			return
		}
	}

}
