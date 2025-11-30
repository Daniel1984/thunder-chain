package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"com.perkunas/internal/models/peernode"
)

type BootstrapNode struct {
	log       *slog.Logger
	port      string
	peerModel *peernode.Model
}

func (bn *BootstrapNode) cleanupInactivePeers(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Println("running periodic cleanup task")
			if _, err := bn.peerModel.MarkInactive(ctx); err != nil {
				bn.log.Error("failed to cleanup inactive peers", "err", err)
			}
		}
	}
}
