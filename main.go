package main

import (
    "fmt"
    "github.com/gorilla/websocket"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

type M map[string]interface{}

const MESSAGE_NEW_USER = "User Baru"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

var connections = make([]*WebSocketConnection, 0)

type Id struct{
	ID int
}

type SocketPayload struct {
     Message string
 }
type SocketResponse struct {
     From    string
     Type    string
     Message string
 }
type WebSocketConnection struct {
     *websocket.Conn
     Username string
 }

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        content, err := ioutil.ReadFile("home.html")
        if err != nil {
            http.Error(w, "Tidak bisa membuka file", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "%s", content)
    })

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
    if err != nil {
        http.Error(w, "Tidak bisa terhubung dengan websocket", http.StatusBadRequest)
    }

    username := r.URL.Query().Get("username")
    currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
    connections = append(connections, &currentConn)

    go handleIO(&currentConn, connections)
    })

    fmt.Println("Buka localhost:8080 pada browser anda")
    http.ListenAndServe(":8080", nil)
}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("ERROR", fmt.Sprintf("%v", r))
        }
    }()

    broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

    for {
        payload := SocketPayload{}
        err := currentConn.ReadJSON(&payload)
        if err != nil {
            if strings.Contains(err.Error(), "websocket: Tutup") {
                broadcastMessage(currentConn, MESSAGE_LEAVE, "")
                return
            }

            log.Println("ERROR", err.Error())
            continue
        }

        broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
    }
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
     for _, eachConn := range connections {
         if eachConn == currentConn {
             continue
         }

         eachConn.WriteJSON(SocketResponse{
             From:    currentConn.Username,
             Type:    kind,
             Message: message,
         })
     }
}