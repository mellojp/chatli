package api

import (
	"net/http"
	"os"

	"github.com/mellojp/chatli/data"

	"github.com/gorilla/websocket"
)

func getWSURL() string {
	val := os.Getenv("WS_URL")
	if val == "" {
		return "ws://localhost:8080"
	}
	return val
}

func ConnectWebSocket(s data.Session) (*websocket.Conn, error) {
	connUrl := getWSURL() + "/ws"
	
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+s.Token)

	socket, _, err := websocket.DefaultDialer.Dial(connUrl, headers)
	if err != nil {
		return nil, err
	}
	return socket, nil
}
