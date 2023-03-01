package websockify

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gidoBOSSftw5731/log"

	"github.com/gorilla/websocket"
)

type Websockify struct {
	Target string
}

// WS is the main function of the websockify package
// It handles the websocket connection and proxies it to the target
// It returns an error if something goes wrong
// Set the target in the websockify struct, this is for compatibility with http.HandleFunc
func (c Websockify) WS(w http.ResponseWriter, r *http.Request) (err error) {
	// Upgrade connection
	upgrader := websocket.Upgrader{Subprotocols: []string{"binary"}, CheckOrigin: originCheckFunc}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	log.Debugf("Rec connection from %s", conn.RemoteAddr())
	defer log.Debugf("Close connection from %s", conn.RemoteAddr())
	defer conn.Close()
	tcpconn, err := net.Dial("tcp", c.Target)
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

// WSNoErr is a wrapper for WS that logs errors instead of returning them
// This is useful for http.HandleFunc and should never be used in any other situation
// This will NOT panic on errors, which may be undesirable
func (c Websockify) WSNoErr(w http.ResponseWriter, r *http.Request) {
	err := c.WS(w, r)
	if err != nil {
		log.Errorln(err)
	}
}

func originCheckFunc(r *http.Request) bool {
	log.Tracef("Origin: %v Host: %v", r.Header.Get("Origin"), r.Host)
	return true
}
