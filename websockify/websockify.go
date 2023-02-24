package websockify

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gidoBOSSftw5731/log"

	"github.com/gorilla/websocket"
)

func WS(w http.ResponseWriter, r *http.Request, target string) (err error) {
	// Upgrade connection
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	log.Debugf("Rec connection from %s", conn.RemoteAddr())
	defer log.Debugf("Close connection from %s", conn.RemoteAddr())
	defer conn.Close()
	tcpconn, err := net.Dial("tcp", target)
	if err != nil {
		return err
	}
	defer tcpconn.Close()
	go func() {
		defer log.Debugf("Close connection from %s", conn.RemoteAddr())
		defer conn.Close()
		defer tcpconn.Close()
		for {
			buffer := make([]byte, 1024)
			n, err := tcpconn.Read(buffer)
			if err != nil || n == 0 {
				return
			}
			conn.WriteMessage(websocket.BinaryMessage, buffer[:n])
		}
	}()
	// Read messages from socket
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		// Is this supposed to be an error?
		if msgType != websocket.BinaryMessage {
			log.Traceln("Non binary message recieved")
		}
		n, err := tcpconn.Write(msg)
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("errWriteEmpty")
		}
	}
}
