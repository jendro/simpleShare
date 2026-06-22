package main

import (
    "flag"
    "log"
    "net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func main() {
    flag.Parse()
    hub := newHub()

    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/ws", wsHandler(hub))
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

    log.Printf("starting server at %s", *addr)
    if err := http.ListenAndServe(*addr, nil); err != nil {
        log.Fatal(err)
    }
}
