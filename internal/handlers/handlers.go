package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"ws/internal/connection"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WebSocket Connection이 보낸 Message에 대한 채널
var wsChannels = make(chan connection.WsPayload)

// 서버에 연결된 Clients
var clients = make(map[connection.WebSocketConnection]string)

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)

	if err != nil {
		log.Println(err)
	}
}

func renderPage(w http.ResponseWriter, templateName string, data jet.VarMap) error {
	view, err := views.GetTemplate(templateName)

	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)

	if err != nil {
		log.Println(err)
	}

	return err
}

func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to endpoint")

	var response connection.WsJsonResponse
	response.Message = `<em><small>Connected to server</small></em>`

	connection := connection.WebSocketConnection{Conn: ws}
	clients[connection] = ""

	err = ws.WriteJSON(response)

	if err != nil {
		log.Println(err)
	}

	go ListenForWS(&connection)
}

// 연결된 Connection이 보내는 Message를 처리한다
func ListenForWS(conn *connection.WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload connection.WsPayload

	for {
		// Connection이 보낸 Message를 받는다
		err := conn.ReadJSON(&payload)

		if err != nil {
			// do nothing
		} else {
			payload.Conn = *conn
			wsChannels <- payload
		}
	}
}

// wsChannels 채널 변경에 대한 처리
func ListenToWsChannel() {
	for {
		e := <-wsChannels

		var response connection.WsJsonResponse
		response.Action = "Got here"
		response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
		boardcastToAll(response)
	}
}

// 모든 연결된 Connection에게 Message를 보낸다
func boardcastToAll(response connection.WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)

		if err != nil {
			log.Println("websocket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}
