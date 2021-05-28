package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/whoisixwebsocket/connection"
)

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	//	w.Write([]byte("hello"))
	var (
		wsConn *websocket.Conn
		err    error
		conn   *connection.Connection
		data   []byte
	)

	// 完成ws协议的握手操作
	// Upgrade:websocket
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	wsConn.SetCloseHandler(func(code int, text string) error {
		log.Println("close from client")
		return wsConn.Close()
	})

	if conn, err = connection.InitConnection(wsConn); err != nil {
		goto ERR
	}

	// 心跳
	go func() {
		var (
			err error
		)
		for {
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		if data, err = conn.ReadMessage(); err != nil {
			goto ERR
		}
		log.Printf("rec:%s\n", data)

		bf := bytes.Buffer{}
		bf.Write(data)
		bf.WriteString("!")
		if err = conn.WriteMessage(bf.Bytes()); err != nil {
			goto ERR
		}
	}

ERR:
	conn.Close()

}

func main() {

	http.HandleFunc("/ws", wsHandler)
	if err := http.ListenAndServe("0.0.0.0:7777", nil); err != nil {
		log.Fatal(err)
	}
}
