package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"com.perkunas/internal/models/transaction"
	"github.com/ethereum/go-ethereum/crypto"
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

	// Parse private key
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid private key format: %v\n", err)
		os.Exit(1)
	}

	privKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid private key: %v\n", err)
		os.Exit(1)
	}

	// Verify that the from address matches the private key
	expectedAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	if from != expectedAddress {
		fmt.Fprintf(os.Stderr, "Error: from address %s does not match private key address %s\n", from, expectedAddress)
		os.Exit(1)
	}

	// Create transaction
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
	if err := transaction.SignTransaction(tx, privKey); err != nil {
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
	// Generate a new private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	// Get the public key address
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Get the private key as hex string
	privateKeyHex := hex.EncodeToString(crypto.FromECDSA(privateKey))

	// Output the key pair
	fmt.Println("Generated Key Pair:")
	fmt.Printf("Private Key: %s\n", privateKeyHex)
	fmt.Printf("Address:     %s\n", address)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
