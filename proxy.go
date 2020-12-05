package main

import (
	"io"
	"log"

	"github.com/gorilla/websocket"
)

func fromClient(c *websocket.Conn, w io.Writer) {
	var (
		buf []byte
		e   error
	)

	buf = make([]byte, 1024)

	for _, buf, e = c.ReadMessage(); e == nil; _, buf, e = c.ReadMessage() {
		if _, e = w.Write(buf); e != nil {
			log.Println(e)
			return
		}
	}
	log.Println(e)
	return
}

func toClient(c *websocket.Conn, r io.Reader) {
	var (
		n   int
		buf []byte
		e   error
	)

	buf = make([]byte, 1024)

	for n, e = r.Read(buf); e == nil && n > 0; n, e = r.Read(buf) {
		c.WriteMessage(websocket.BinaryMessage, buf[:n])
	}
	log.Println(e)
	return
}
