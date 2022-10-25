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

type Message struct {
	MsgType string
	Msg     map[string]interface{}
}

var Clients = make(map[*websocket.Conn]bool)
var Broadcast = make(chan Message)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func CheckLoginInfo(userid interface{}, userpw interface{}) int {

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

/*
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
*/
func socketHandler(w http.ResponseWriter, r *http.Request) {

	var res string

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
	fmt.Println("new connection created")

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

		/*
			json 메시지를 decoding 한다.
		*/
		data := Message{}
		json.Unmarshal([]byte(p), &data)
		fmt.Println(string(p))

		if data.MsgType == "LOGIN" {
			/*
				LOGIN 메시지이면 ID, PW 입력하여 로그인한다.
			*/
			fmt.Println(data.Msg["userid"], data.Msg["userpw"])
			if CheckLoginInfo(data.Msg["userid"], data.Msg["userpw"]) < 0 {
				fmt.Println("Login fail")
				res = "LOGIN_FAIL"
			} else {
				fmt.Println("Login Success")
				res = "LOGIN_SUCC"
			}

			/*
				결과를 클라이언트로 전송한다.
			*/
			bytedata, _ := json.Marshal(res)
			if err := conn.WriteMessage(msgType, bytedata); err != nil {
				log.Printf("conn.WriteMessage : %v\n", err)
				delete(Clients, conn)
				return
			}
		} else if data.MsgType == "MESSAGE" {
			/*
				MESSAGE 이면 연결된 클라이언트 전체에게 메시지를 전달한다.
			*/
			fmt.Printf("Msg send :: %s\n", data.Msg["msg"])

			data_msg, _ := json.Marshal(data.Msg["msg"])
			for ClientConn := range Clients {
				if err := ClientConn.WriteMessage(msgType, data_msg); err != nil {
					log.Printf("conn.WriteMessage : %v\n", err)
					delete(Clients, ClientConn)
					return
				}
			}
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
	//go BroadcastHandler()

	/*
		설정 포트로 웹 서버를 실행한다.
	*/
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe : ", err)
	}
}
