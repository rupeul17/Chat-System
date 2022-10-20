package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %v\n", err)
		return
	}

	defer conn.Close()

	for {
		msgType, p, err := conn.ReadMessage()
		fmt.Println(string(p))

		if err != nil {
			log.Printf("conn.ReadMessage : %v\n", err)
			return
		}

		if err := conn.WriteMessage(msgType, p); err != nil {
			log.Printf("conn.WriteMessage : %v\n", err)
			return
		}
	}
}

func main() {

	http.Handle("/", http.FileServer(http.Dir("page")))
	http.HandleFunc("/ws", socketHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe : ", err)
	}
}
