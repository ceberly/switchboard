package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var switchboard = NewSwitchboard()

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
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := switchboard.Subscribe(1)
	defer switchboard.Unsubscribe(1, ch)

	for {
		data := <-ch
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
	log.Println(r)
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&p)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("publishing %v", p)
	switchboard.Publish(1, p)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "OK")
}

func main() {
	log.Println("starting up")
	http.HandleFunc("/subscribe", subscribe)
	http.HandleFunc("/publish", publish)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
