package main

import (
	"context"

	"com.perkunas/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) GetNodeStatus(ctx context.Context, req *proto.GetNodeStatusRequest) (*proto.NodeStatusResponse, error) {
	latestBlock, err := n.blocksRPC.GetLatestBlock(ctx, nil)
	if err != nil {
		n.log.Error("failed gettin latest block", "err", err)
		return nil, status.Error(codes.Internal, "failed gettin latest block")
	}

	n.log.Info("latestBlock", latestBlock)

	return &proto.NodeStatusResponse{}, nil
}
