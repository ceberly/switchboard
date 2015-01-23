package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var (
	switchboard = NewSwitchboard()
	subscribeRE = regexp.MustCompile("^/subscribe/([0-9]+)/?$")
	publishRE = regexp.MustCompile("^/publish/([0-9]+)/?$")
)

func subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m := subscribeRE.FindStringSubmatch(r.URL.Path)
	if len(m) != 2 {
		http.Error(w, "Bad request", http.StatusBadRequest)
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

	ch := switchboard.Subscribe(ChannelIndex(m[1]))
	defer switchboard.Unsubscribe(ChannelIndex(m[1]), ch)

	closeNotifier, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-closeNotifier.CloseNotify():
			return
		}
	}
}

func publish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m := publishRE.FindStringSubmatch(r.URL.Path)
	if len(m) != 2 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	switchboard.Publish(ChannelIndex(m[1]), body)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "OK")
}

func main() {
	log.Println("starting up")
	http.HandleFunc("/subscribe/", subscribe)
	http.HandleFunc("/publish/", publish)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
