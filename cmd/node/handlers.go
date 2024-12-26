package main

import (
	"encoding/json"
	"net/http"
	"time"

	"com.perkunas/internal/httpjsonres"
	"com.perkunas/internal/models/transaction"
	"com.perkunas/proto"
)

func (a *App) createTransaction(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var txn transaction.Transaction
	if err := json.NewDecoder(r.Body).Decode(&txn); err != nil {
		a.log.Error("could not read request body", "err", err)
		http.Error(w, "could not create transaction", http.StatusBadRequest)
		return
	}

	if txn.Timestamp == 0 {
		txn.Timestamp = time.Now().Unix()
	}

	if txn.Expires == 0 {
		txn.Expires = time.Now().Add(15 * time.Minute).Unix()
	}

	if err := txn.Verify(); err != nil {
		a.log.Error("invalid or tampered transaction", "err", err)
		http.Error(w, "invalid or tampered transaction", http.StatusBadRequest)
		return
	}

	protoTxn := proto.Transaction{
		Hash:      txn.Hash,
		FromAddr:  txn.From,
		ToAddr:    txn.To,
		Signature: txn.Signature,
		Amount:    txn.Amount,
		Fee:       txn.Fee,
		Timestamp: txn.Timestamp,
		Expires:   txn.Expires,
	}

	pld := &proto.CreateTransactionRequest{Transaction: &protoTxn}
	createResp, err := a.rpcClient.CreateTransaction(r.Context(), pld)
	if err != nil {
		a.log.Error("could not create transaction", "err", err)
		http.Error(w, "could not create transaction", http.StatusBadRequest)
		return
	}

	if err := httpjsonres.JSON(w, http.StatusOK, createResp); err != nil {
		a.log.Error("failed responding to create transaction request", "err", err)
	}
}
