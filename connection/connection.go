package connection

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	wsConnect *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte

	once sync.Once
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConnect: wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}
	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()
	return
}

func (conn *Connection) ReadMessage() (data []byte, err error) {

	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func (conn *Connection) WriteMessage(data []byte) (err error) {

	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}
	return
}

func (conn *Connection) Close() {
	// 线程安全，可多次调用
	conn.wsConnect.Close()

	conn.once.Do(func() {
		close(conn.closeChan)
	})
}

// 内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	defer conn.Close()

	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			log.Println("ReadMessage err", err)
			return
		}
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn断开
			return
		}

	}

}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	defer conn.Close()

	select {
	case data = <-conn.outChan:
	case <-conn.closeChan:
		return
	}

	for {
		data = <-conn.outChan
		if err = conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}
