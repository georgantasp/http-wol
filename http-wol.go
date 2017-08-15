package main

import (
  "os"
  "io"
  "net/http"
  "os/exec"
)

func wol(w http.ResponseWriter, r *http.Request) {
  out, err := exec.Command("/usr/bin/wakeonlan", r.URL.Query().Get("mac")).Output()
  if err != nil {
          io.WriteString(w, "An error occurred")
  } else {
          io.WriteString(w, string(out[:]))
  }
}

func main() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", wol)
  http.ListenAndServe(":" + os.Args[1], mux)
}
