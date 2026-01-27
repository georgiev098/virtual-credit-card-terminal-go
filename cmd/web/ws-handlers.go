package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketConn struct {
	*websocket.Conn
}

type Payload struct {
	Action      string        `json:"action"`
	Message     string        `json:"message"`
	UserName    string        `json:"username"`
	MessageType string        `json:"message_type"`
	Connection  WebSocketConn `json:"-"`
}

type WsJSONResp struct {
	Action  string `json:"action"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

var upgradeConn = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[WebSocketConn]string, 0)

var wsChan = make(chan Payload)

func (app *Application) WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConn.Upgrade(w, r, nil)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	app.InfoLog.Printf(fmt.Sprintf("Client connected from %s", r.RemoteAddr))

	var resp WsJSONResp

	resp.Message = "Connected to server."

	err = ws.WriteJSON(resp)
	if err != nil {
		app.ErrorLog.Println(err)
		return
	}

	conn := WebSocketConn{
		Conn: ws,
	}

	clients[conn] = ""

	go app.ListenForWs(&conn)
}

func (app *Application) ListenForWs(conn *WebSocketConn) {
	defer func() {
		if r := recover(); r != nil {
			app.ErrorLog.Println("Error: " + fmt.Sprintf("%v", r))
		}
	}()

	var payload Payload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
		} else {
			payload.Connection = *conn
			wsChan <- payload
		}
	}
}

func (app *Application) ListenToWsChan() {
	var resp WsJSONResp

	for {
		e := <-wsChan
		switch e.Action {
		case "deleteUser":
			resp.Action = "logout"
			resp.Message = "Your account has been deleted."
			app.BroadcastToAll(resp)
		default:
		}
	}
}

func (app *Application) BroadcastToAll(resp WsJSONResp) {
	for client := range clients {
		// bradcaast to every client
		err := client.WriteJSON(resp)
		if err != nil {
			app.ErrorLog.Printf("Websocket error on %s, %s", resp.Action, err)
			_ = client.Close()
			delete(clients, client)
		}

	}
}
