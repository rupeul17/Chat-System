package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type ClientMsgInfo struct {
	MsgType string
	UserId  string
	UserPw  string
	Msg     []byte
}

type Message struct {
	MsgType int
	MsgInfo ClientMsgInfo
}

var Clients = make(map[*websocket.Conn]bool)
var Broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func BroadcastHandler() {

	for {
		select {
		case Msg, ok := <-Broadcast:
			if ok {
				for conn := range Clients {
					if err := conn.WriteMessage(Msg.MsgType, Msg.MsgInfo.Msg); err != nil {
						log.Printf("conn.WriteMessage : %v\n", err)
						return
					}
				}
			}
		}
	}
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	/*
		요청 및 응답을 http -> ws 로 변경한다.
	*/
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %v\n", err)
		return
	}

	/*
		함수 종료 시 conn을 종료한다.
	*/
	defer conn.Close()

	Clients[conn] = true

	for {
		/*
			클라이언트 수신 메시지를 읽고 출력한다.
		*/
		msgType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage : %v\n", err)
		}

		data := ClientMsgInfo{}
		json.Unmarshal([]byte(p), &data)
		fmt.Println(string(p))

		Msg := Message{
			MsgType: msgType,
		}

		if data.MsgType == "LOGIN" {
			Msg.MsgInfo.Msg = p
			Broadcast <- Msg
		} else {
			Msg.MsgInfo.Msg = p
			Broadcast <- Msg
		}
	}
}

func main() {

	/*
		설정한 url 에 접근했을 때 설정한 파일 서버를 동작시킨다.
	*/
	http.Handle("/", http.FileServer(http.Dir("login")))
	/*
		ws는 websocket의 약자
	*/
	http.HandleFunc("/ws", socketHandler)

	/*
		broadcast 고루틴 실행
	*/
	go BroadcastHandler()

	/*
		설정 포트로 웹 서버를 실행한다.
	*/
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe : ", err)
	}
}
