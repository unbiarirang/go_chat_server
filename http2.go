package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type client chan string

type Room struct {
	members  map[client]bool
	enter    chan client
	messages chan string
}

func (r *Room) serve() {
	for {
		select {
		case client := <-r.enter:
			r.members[client] = true
		case message := <-r.messages:
			r.broadCast(message)
		}
	}
}

func (r *Room) broadCast(message string) {
	for client := range lobby.members {
		client <- message
	}
}

var lobby = Room{
	members:  make(map[client]bool),
	enter:    make(chan client),
	messages: make(chan string),
}

const loginHTML = `<html>
	<head><title>Welcome to CHATTING GO</title>
	</head>
	<body>
	<form action="/clockstream">
	Nickname:<br>
	<input type="text" name="nick">
	<br>
	<input type="submit" value="Submit">
	</form>
	</body>
	</html>`

func main() {
	go lobby.serve()

	var srv http.Server
	srv.Addr = "localhost:5000"

	http.HandleFunc("/clockstream", clockStreamHandler)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, loginHTML)
	})
	http.HandleFunc("/", chatHandler)

	http2.ConfigureServer(&srv, &http2.Server{})
	log.Fatal(http.ListenAndServeTLS(":5000", "cert.pem", "key.pem", nil))

	// http.ListenAndServe(":5000", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	message := r.URL.Path[1:]
	//if message != "/favicon.ico" {
	lobby.messages <- fmt.Sprintf("%v: %v", nick, message)
	//}
}

func clockStreamHandler(w http.ResponseWriter, r *http.Request) {
	client := make(chan string)
	var nick = r.URL.Query().Get("nick")

	go func() { lobby.enter <- client }()
	go func() { lobby.messages <- "*****" + nick + " entered*****" }()

	clientGone := w.(http.CloseNotifier).CloseNotify()
	w.Header().Set("Content-Type", "text/plain")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Fprintf(w, "# ~1KB of junk to force browsers to start rendering immediately: \n")
	io.WriteString(w, strings.Repeat("# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 13))
	for {
		select {
		case msg := <-client:
			fmt.Fprintf(w, "%s\n", msg)
			fmt.Fprintf(w, "%v\n", time.Now())
			w.(http.Flusher).Flush()
		case <-clientGone:
			log.Printf("Client %v disconnected from the clock", r.RemoteAddr)
			lobby.messages <- "*****" + nick + " left*****"
			delete(lobby.members, client)
			//채팅 종료 코드
			return
		}
	}
}
