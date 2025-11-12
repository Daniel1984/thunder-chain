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
		case t := <-ticker.C:
			fmt.Println("running periodic task at", t.Format(time.RFC3339))
			res, err := bn.peerModel.MarkInactive(ctx)
			if err != nil {
				bn.log.Error("failed to cleanup inactive peers", "err", err)
				return
			}

			if rowsAffected, err := res.RowsAffected(); err == nil && rowsAffected > 0 {
				bn.log.Debug("marked peers as inactive", "count", rowsAffected)
			}
		}
	}
}
