package main

import (
	"encoding/json"
	"net/http"
	"time"

	"com.perkunas/internal/httpjsonres"
	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
)

func (n *Node) createTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

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
		txn.Expires = time.Now().Add(15 * time.Minute).Unix()
	}

	if err := txn.Verify(); err != nil {
		n.log.Error("invalid or tampered transaction", "tx", txn, "err", err)
		http.Error(w, "invalid or tampered transaction data", http.StatusBadRequest)
		return
	}

	protoTxn := transaction.ToProtoTx(txn)
	pld := &proto.CreateMempoolRequest{Transaction: protoTxn}
	createResp, err := n.mempoolRPC.CreateMempool(r.Context(), pld)
	if err != nil {
		n.log.Error("could not push transaction to mempool", "txHash", txn.Hash, "err", err)
		http.Error(w, "could not create transaction", http.StatusBadRequest)
		return
	}

	if err := httpjsonres.JSON(w, http.StatusOK, createResp); err != nil {
		n.log.Error("failed responding to create transaction request", "err", err)
	}
}

func (n *Node) nodeStatus(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	lb, err := n.stateRPC.GetLatestBlock(r.Context(), &proto.LastBlockReq{})
	if err != nil {
		n.log.Error("could not get latest block", "err", err)
		http.Error(w, "could not get latest block", http.StatusBadRequest)
		return
	}

	if err := httpjsonres.JSON(w, http.StatusOK, lb); err != nil {
		n.log.Error("failed responding to get latest block from stet service", "err", err)
	}
}
