package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"com.perkunas/internal/models/transaction"
	"com.perkunas/pkg/wallet"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "perkunas-cli",
	Short: "Perkunas chain CLI tool",
	Long:  "A command line interface for interacting with the Perkunas chain",
}

var signTxCmd = &cobra.Command{
	Use:   "sign-tx",
	Short: "Sign a transaction",
	Long:  "Sign a blockchain transaction with the provided parameters",
	Run:   signTransaction,
}

var generateKeysCmd = &cobra.Command{
	Use:   "generate-keys",
	Short: "Generate a private/public key pair",
	Long:  "Generate a new private key and corresponding public key address",
	Run:   generateKeys,
}

// Command line flags
var (
	from       string
	to         string
	amount     int64
	fee        int64
	nonce      uint64
	privateKey string
	data       string
)

func init() {
	// Add flags to the sign-tx command
	signTxCmd.Flags().StringVarP(&from, "from", "f", "", "Sender address (required)")
	signTxCmd.Flags().StringVarP(&to, "to", "t", "", "Recipient address (required)")
	signTxCmd.Flags().Int64VarP(&amount, "amount", "a", 0, "Amount to transfer (required)")
	signTxCmd.Flags().Int64Var(&fee, "fee", 0, "Transaction fee (required)")
	signTxCmd.Flags().Uint64VarP(&nonce, "nonce", "n", 0, "Transaction nonce (required)")
	signTxCmd.Flags().StringVarP(&privateKey, "private-key", "k", "", "Private key for signing (hex format, required)")
	signTxCmd.Flags().StringVarP(&data, "data", "d", "", "Additional transaction data (optional)")

	// Mark required flags
	signTxCmd.MarkFlagRequired("from")
	signTxCmd.MarkFlagRequired("to")
	signTxCmd.MarkFlagRequired("amount")
	signTxCmd.MarkFlagRequired("fee")
	signTxCmd.MarkFlagRequired("nonce")
	signTxCmd.MarkFlagRequired("private-key")

	// Add commands to root
	rootCmd.AddCommand(signTxCmd)
	rootCmd.AddCommand(generateKeysCmd)
}

func signTransaction(cmd *cobra.Command, args []string) {
	// Validate inputs
	if amount < 0 {
		fmt.Fprintf(os.Stderr, "Error: amount cannot be negative\n")
		os.Exit(1)
	}

	if fee < 0 {
		fmt.Fprintf(os.Stderr, "Error: fee cannot be negative\n")
		os.Exit(1)
	}

	// Create wallet from private key
	w, err := wallet.FromPrivateKey(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Verify that the from address matches the wallet address
	if from != w.Address {
		fmt.Fprintf(os.Stderr, "Error: from address %s does not match wallet address %s\n", from, w.Address)
		os.Exit(1)
	}

	tx := &transaction.Transaction{
		From:      from,
		To:        to,
		Amount:    amount,
		Fee:       fee,
		Nonce:     nonce,
		Data:      data,
		Timestamp: time.Now().Unix(),
		Expires:   time.Now().Add(24 * time.Hour).Unix(), // Default 24 hour expiry
	}

	// Set transaction hash
	tx.SetHash()

	// Sign transaction
	if err := w.SignTransaction(tx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to sign transaction: %v\n", err)
		os.Exit(1)
	}

	// Verify signature (optional check)
	if err := tx.Verify(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: transaction verification failed: %v\n", err)
		os.Exit(1)
	}

	// Output signed transaction as JSON
	jsonOutput, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal transaction to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonOutput))
}

func generateKeys(cmd *cobra.Command, args []string) {
	w, err := wallet.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to generate wallet: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated Key Pair:")
	fmt.Printf("Private Key: %s\n", w.GetPrivateKeyHex())
	fmt.Printf("Address:     %s\n", w.Address)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
