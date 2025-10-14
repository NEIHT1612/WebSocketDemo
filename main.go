package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

// convert connection from http to websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// save all connections websocket is active
var mapWsConn = make(map[string]*websocket.Conn)

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/chat", LoadPageChat)
	http.HandleFunc("/ws", InitWebsocket)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// read html file and load to browser
func LoadPageChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	path, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(w, "%s", "error")
		return
	}

	content, err := os.ReadFile(path + "/chat-using-websocket/chat.html")
	if err != nil {
		fmt.Fprintf(w, "%s", "error")
		return
	}

	fmt.Fprintf(w, "%s", content)
}

func InitWebsocket(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	channel := r.URL.Query().Get("channel")
	if r.Header.Get("Origin") != "http://"+r.Host {
		fmt.Fprintf(w, "%s", "error")
		return
	}

	if _, ok := mapWsConn[channel]; !ok {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Fprintf(w, "%s", "error")
			return
		}

		mapWsConn[channel] = conn
	}

	for {
		var msg map[string]string
		err := mapWsConn[channel].ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading JSON: ", err)
			break
		}
		fmt.Printf("Received: %s\n", msg)

		otherConn := getConn(channel)
		if otherConn == nil {
			continue
		}

		err = otherConn.WriteJSON(msg)
		if err != nil {
			fmt.Println("Error writing JSON: ", err)
			break
		}
	}
}

func getConn(channel string) *websocket.Conn {
	for key, conn := range mapWsConn {
		if key != channel {
			return conn
		}
	}
	return nil
}
