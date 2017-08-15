package main

import (
  "os"
  "fmt"
  "net/http"
  wol "github.com/sabhiram/go-wol"
)

func wol(w http.ResponseWriter, r *http.Request) {
  macAddr = r.URL.Query().Get("mac")
  
  err = wol.SendMagicPacket(macAddr, "255.255.255.255:9", "")
  
  if err != nil {
    fmt.Fprintf(w, "ERROR: %s\n", err.Error())
  } else {
    fmt.Fprintf("Magic packet sent successfully to %s\n", macAddr)
  }
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", wol)
  http.ListenAndServe(":" + os.Args[1], mux)
}
