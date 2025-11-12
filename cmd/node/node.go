package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

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
}

type RegisterPeerRequest struct {
	Port string `json:"port"`
}

func (n *Node) RegisterAsPeer(ctx context.Context) error {
	ctxt, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	payload := RegisterPeerRequest{
		Port: n.apiPort,
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
		return fmt.Errorf("bootstrap returned status %d", resp.StatusCode)
	}

	return nil
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
