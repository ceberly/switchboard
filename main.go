package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Post struct {
	id         int
	ramble_id  int
	text       string
	user_id    int
	created_at string
	updated_at string
}

var subscriptions = make(map[chan []byte]bool)

func subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "This connection does not support server side events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	subscription := make(chan []byte)
	//XXX: I don't think that map access is atomic
	subscriptions[subscription] = true

	for {
		data := <-subscription
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

func publish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p Post
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(p)
	log.Println("publshing ", b)

	for s := range subscriptions {
		s <- b
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "OK")
}

func main() {
	http.HandleFunc("/subscribe", subscribe)
	http.HandleFunc("/publish", publish)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
