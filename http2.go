package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var c = make(map[int]chan string)
var i = 0

func main() {
	http.HandleFunc("/clockstream", clockStreamHandler)
	http.HandleFunc("/", chatHandler)

	http.ListenAndServe(":5000", nil)
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	for _, ch := range c {
		if r.URL.Path != "/favicon.ico" {
			ch <- r.URL.Path
		}
	}
}

func clockStreamHandler(w http.ResponseWriter, r *http.Request) {
	var ch = make(chan string)
	var index = i
	c[index] = ch
	i++
	fmt.Println(c)
	clientGone := w.(http.CloseNotifier).CloseNotify()
	w.Header().Set("Content-Type", "text/plain")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	fmt.Fprintf(w, "# ~1KB of junk to force browsers to start rendering immediately: \n")
	io.WriteString(w, strings.Repeat("# xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\n", 13))
	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "%s\n", msg)
			fmt.Fprintf(w, "%v\n", time.Now())
			w.(http.Flusher).Flush()
		case <-clientGone:
			log.Printf("Client %v disconnected from the clock", r.RemoteAddr)
			delete(c, index)
			fmt.Println(c)
			return
		}
	}
}
