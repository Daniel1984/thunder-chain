package main

import (
	"net/http"
)

func (n *Node) getRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /transactions", n.createTransaction)
	mux.HandleFunc("GET /status", n.nodeStatus)

	return mux
}
