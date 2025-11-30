package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"com.perkunas/internal/models/peernode"
	"com.perkunas/proto"
)

type Node struct {
	log                        *slog.Logger
	apiPort                    string
	mempoolAPI                 string
	stateAPI                   string
	bootstrapAPI               string
	participateInPeerDiscovery bool
	mempoolRPC                 proto.MempoolServiceClient
	stateRPC                   proto.StateServiceClient
	peerNodes                  []peernode.Peer
	outboundIP                 string
	mutex                      sync.RWMutex
}

type RegisterPeerRequest struct {
	Port string `json:"port"`
	IP   string `json:"ip"`
}

func (n *Node) RegisterAsPeer(ctx context.Context) error {
	ctxt, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	payload := RegisterPeerRequest{
		Port: n.apiPort,
		IP:   n.outboundIP,
	}

	jsonData, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/register", n.bootstrapAPI)
	req, _ := http.NewRequestWithContext(ctxt, "POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RegisterAsPeer returned status %d", resp.StatusCode)
	}

	return nil
}

func (n *Node) FetchPeers(ctx context.Context) {
	ctxt, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/peers", n.bootstrapAPI)
	req, _ := http.NewRequestWithContext(ctxt, "GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		n.log.Error("failed to fetch peers", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		n.log.Error("FetchPeers failed", "status code", resp.StatusCode)
		return
	}

	var peers []peernode.Peer
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		n.log.Error("FetchPeers failed decoding response", "err", err)
		return
	}

	if len(peers) == 0 {
		return
	}

	filteredPeers := []peernode.Peer{}
	for _, peer := range peers {
		if peer.IP != n.outboundIP {
			filteredPeers = append(filteredPeers, peer)
		}
	}

	n.peerNodes = filteredPeers
}

func (n *Node) StartHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := n.RegisterAsPeer(ctx); err != nil {
				n.log.Error("heartbeat failed", "err", err)
			}
		}
	}
}

func (n *Node) StartPeerRefresher(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n.FetchPeers(ctx)
		}
	}
}

func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}
