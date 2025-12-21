package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"time"

	"com.perkunas/internal/httpjsonres"
	"com.perkunas/internal/models/peernode"
	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
)

func (n *Node) createTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var txn transaction.Transaction
	if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
		n.log.Error("could not read request body", "err", err)
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if txn.Timestamp == 0 {
		txn.Timestamp = time.Now().Unix()
	}

	if txn.Expires == 0 {
		txn.Expires = time.Now().Add(10 * time.Minute).Unix()
	}

	if err := txn.Verify(); err != nil {
		n.log.Error("invalid or tampered transaction", "tx", txn, "err", err)
		http.Error(w, "invalid or tampered transaction data", http.StatusBadRequest)
		return
	}

	fromAcc, err := n.stateRPC.GetAccountByAddress(ctx, &proto.AccountByAddressReq{Address: txn.From})
	if err != nil {
		n.log.Error("could not get account by address", "err", err)
		http.Error(w, "could not get account by address", http.StatusBadRequest)
		return
	}

	accNonce := fromAcc.GetAccount().GetNonce()
	if txn.Nonce != accNonce+1 {
		n.log.Warn("invalid nonce", "acc nonce", accNonce, "tx nonce", txn.Nonce)
		http.Error(w, "invalid nonce", http.StatusBadRequest)
		return
	}

	protoTxn := transaction.ToProtoTx(txn)
	pld := &proto.CreateMempoolRequest{Transaction: protoTxn}
	createResp, err := n.mempoolRPC.CreateMempool(ctx, pld)
	if err != nil {
		n.log.Error("could not push transaction to mempool", "txHash", txn.Hash, "err", err)
		http.Error(w, "could not create transaction", http.StatusBadRequest)
		return
	}

	// broadcast transaction to peer nodes after successful mempool storage
	go n.broadcastTransaction(txn)

	if err := httpjsonres.JSON(w, http.StatusOK, createResp); err != nil {
		n.log.Error("failed responding to create transaction request", "err", err)
	}
}

func (n *Node) nodeStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	lb, err := n.stateRPC.GetLatestBlock(ctx, &proto.LastBlockReq{})
	if err != nil {
		n.log.Error("could not get latest block", "err", err)
		http.Error(w, "could not get latest block", http.StatusBadRequest)
		return
	}

	if err := httpjsonres.JSON(w, http.StatusOK, lb); err != nil {
		n.log.Error("failed responding to get latest block from state service", "err", err)
	}
}

func (n *Node) broadcastTransaction(txn transaction.Transaction) {
	n.mutex.RLock()
	peers := append([]peernode.Peer(nil), n.peerNodes...)
	n.mutex.RUnlock()

	if len(peers) == 0 {
		n.log.Debug("no peers available", "txHash", txn.Hash)
		return
	}

	client := &http.Client{}
	sem := make(chan struct{}, 16)

	for _, peer := range peers {
		sem <- struct{}{}
		go func(p peernode.Peer) {
			defer func() { <-sem }()
			n.sendTransactionToPeer(p, txn, client)
		}(peer)
	}
}

func (n *Node) sendTransactionToPeer(peer peernode.Peer, txn transaction.Transaction, client *http.Client) {
	ctxt, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	peerAddr := net.JoinHostPort(peer.IP, peer.Port)
	endpoint := "http://" + peerAddr + "/transactions"

	jsonData, err := json.Marshal(txn)
	if err != nil {
		n.log.Error("marshal tx failed", "peer", peerAddr, "err", err)
		return
	}

	req, err := http.NewRequestWithContext(ctxt, http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		n.log.Error("create request failed", "peer", peerAddr, "err", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		n.log.Warn("send tx failed", "peer", peerAddr, "err", err)
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		n.log.Debug("peer rejected tx", "peer", peerAddr, "status", resp.StatusCode)
	}
}
