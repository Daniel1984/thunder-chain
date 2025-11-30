package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"com.perkunas/internal/httpjsonres"
	"com.perkunas/internal/models/peernode"
)

func (bn *BootstrapNode) reportAsPeer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	defer r.Body.Close()

	var pld peernode.Peer
	if err := json.NewDecoder(r.Body).Decode(&pld); err != nil {
		bn.log.Error("invalid request payload", "err", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if pld.Port == "" {
		http.Error(w, "port is required", http.StatusBadRequest)
		return
	}

	pld.IsActive = true
	pld.LastSeen = time.Now().Unix()

	if err := bn.peerModel.Insert(ctx, pld); err != nil {
		bn.log.Error("failed to register peer", "err", err, "ip", pld.IP, "port", pld.Port)
		http.Error(w, "failed to register peer", http.StatusInternalServerError)
		return
	}

	bn.log.Info("peer registered", "ip", pld.IP, "port", pld.Port)
	if err := httpjsonres.JSON(w, http.StatusOK, map[string]string{"status": "OK"}); err != nil {
		bn.log.Error("failed responding to register peer request", "err", err)
	}
}

func (bn *BootstrapNode) getPeers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	defer r.Body.Close()

	peers, err := bn.peerModel.QueryActive(ctx)
	if err != nil {
		bn.log.Error("failed to get active peers", "err", err)
		http.Error(w, "failed to get active peers", http.StatusInternalServerError)
		return
	}

	if err := httpjsonres.JSON(w, http.StatusOK, peers); err != nil {
		bn.log.Error("failed responding to get peers request", "err", err)
	}
}

func (bn *BootstrapNode) healthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	defer r.Body.Close()

	count, err := bn.peerModel.ActiveCount(ctx)
	if err != nil {
		bn.log.Error("failed to count peers", "err", err)
		http.Error(w, "health check failed", http.StatusInternalServerError)
		return
	}

	res := map[string]any{
		"status":     "healthy",
		"peer_count": count,
	}

	if err := httpjsonres.JSON(w, http.StatusOK, res); err != nil {
		bn.log.Error("failed responding to health check request", "err", err)
	}
}
