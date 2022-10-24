package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
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

func CheckLoginInfo(userid string, userpw string) int {

	var UserCount int

	db, err := sql.Open("mysql", "root:1234@tcp(localhost:3306)/member_db")
	if err != nil {
		log.Println(err.Error())
		return -1
	}

	defer db.Close()

	result, err := db.Query("SELECT count(*) FROM member WHERE ID=? AND PASSWD=?", userid, userpw)
	if err != nil {
		log.Println(err.Error())
		return -1
	}

	for result.Next() {
		if err := result.Scan(&UserCount); err != nil {
			log.Fatal(err)
		}
	}

	if UserCount > 0 {
		return 1
	} else {
		return -1
	}
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
			delete(Clients, conn)
			return
		}

		data := ClientMsgInfo{}
		json.Unmarshal([]byte(p), &data)
		fmt.Println(string(p))
		var res string

		if data.MsgType == "LOGIN" {
			if CheckLoginInfo(data.UserId, data.UserPw) < 0 {
				fmt.Println("Login fail")
				res = "LOGIN_FAIL"

			} else {
				fmt.Println("Login Success")
				res = "LOGIN_SUCC"
			}
		} else if data.MsgType == "MESSAGE" {
			fmt.Println("message receive")

			data, _ := json.Marshal(data.Msg)
			for ClientConn := range Clients {
				if err := ClientConn.WriteMessage(msgType, data); err != nil {
					log.Printf("conn.WriteMessage : %v\n", err)
					delete(Clients, ClientConn)
					return
				} else {
					fmt.Println("Msg send")
				}
			}
		}

		bytedata, _ := json.Marshal(res)

		if err := conn.WriteMessage(msgType, bytedata); err != nil {
			log.Printf("conn.WriteMessage : %v\n", err)
			delete(Clients, conn)
			return
		}
		fmt.Printf("res msg send\n")
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
