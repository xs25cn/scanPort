package wsConn

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

func New(wsSocket *websocket.Conn) *WsConnection {

	wsConn := &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *Message, 1000),
		outChan:   make(chan *Message, 1000),
		closeChan: make(chan byte, 1),
	}
	go wsConn.readLoop()
	go wsConn.writeLoop()
	return wsConn
}

type Message struct {
	MessageType int
	Data        []byte
}
type WsConnection struct {
	wsSocket *websocket.Conn
	inChan   chan *Message
	outChan  chan *Message

	mtx       sync.Mutex
	closeChan chan byte
	isClosed  bool
}

func (s *WsConnection) readLoop() {
	for {
		msgType, data, err := s.wsSocket.ReadMessage()
		//fmt.Println(msgType, string(data), err)
		if err != nil {
			s.Close()
		}
		select {
		case s.inChan <- &Message{MessageType: msgType, Data: data}:
			//fmt.Println("接收数据，写入数据到inChan")
		case <-s.closeChan:
			return
		}

	}
}
func (s *WsConnection) writeLoop() {
	for {
		select {
		case msg := <-s.outChan:
			if err := s.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				s.Close()
				return
			}
		case <-s.closeChan:
			return
		}
	}
}

func (s *WsConnection) ReadMessage() (*Message, error) {
	select {
	case msg := <-s.inChan:
		return msg, nil
	case <-s.closeChan:
		return nil, errors.New("websocket closed")
	}
}

func (s *WsConnection) WriteMessage(messageType int, data []byte) error {
	t := time.Now().Format("2006-01-02 15:04:05")
	data = []byte(t + " " + string(data))
	select {
	case s.outChan <- &Message{MessageType: messageType, Data: data}:
		return nil
	case <-s.closeChan:
		return errors.New("websocket closed")
	}
}

//关闭socket连接，关闭通道
func (s *WsConnection) Close() {
	s.wsSocket.Close()
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if !s.isClosed {
		fmt.Println("close")
		close(s.closeChan)
		s.isClosed = true
	}
}
