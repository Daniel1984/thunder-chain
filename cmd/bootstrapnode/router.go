package main

import (
	"net/http"
)

func (bn *BootstrapNode) getRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", bn.reportAsPeer)
	mux.HandleFunc("GET /peers", bn.getPeers)
	mux.HandleFunc("GET /health", bn.healthCheck)

	return mux
}
