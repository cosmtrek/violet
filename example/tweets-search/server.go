package main

import (
  "flag"
  "log"
  "net/http"
)

func main() {
  port := flag.String("p", "8100", "port to serve on")
  directory := flag.String("d", "./dist", "the directory of static file to host")
  flag.Parse()

  http.Handle("/", http.FileServer(http.Dir(*directory)))

  log.Printf("tweets-search: serving %s on HTTP port: %s\n", *directory, *port)
  log.Fatal(http.ListenAndServe(":"+*port, nil))
}
