package main

import "net/http"

func HandleHealth(resp http.ResponseWriter, req *http.Request) {
	data := []byte("OK")
	resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
	resp.WriteHeader(200)
	resp.Write(data)
}
