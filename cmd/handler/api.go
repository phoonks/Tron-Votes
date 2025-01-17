package handler

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	cf "github.com/kahsengphoon/Tron-Votes/config"

	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	shastaDomain  = "grpc.shasta.trongrid.io:50051"
	ownAddress    = "TKjxS5iEAzyJHY9g7ZJR8XwTshDUEeH379"
	hexOwnAddress = "416b2fb069425a03935b9e39d372319cdc1a4fed9f"
	privateKeyHex = "d1455980b1f333aba7a1abf852216abfeec8b95b230b6f5bf1af21c0e4a0c898"
)

func (h *HttpServer) StartApiServer(c *cli.Context) error {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	if h.isStarted {
		return errors.New("Server already started")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	h.isStarted = true
	h.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", h.port),
		Handler: r,
	}

	// freezeBalance()
	// getVotePower()
	getSpAddress()
	// getClaimableReward()
	// voteTransaction()

	if err := r.Run(fmt.Sprintf(":%v", cf.Enviroment().AppServerPort)); err != nil {
		return err
	}

	return nil
}

func freezeBalance() {
	conn := client.NewGrpcClient(shastaDomain)
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		fmt.Printf("fail to NewGrpcClient: %+v \n", conn)
		return
	}
	defer conn.Stop()

	rawTx, err := conn.FreezeBalanceV2(ownAddress, core.ResourceCode_ENERGY, 1000000)
	if err != nil {
		fmt.Printf("fail to FreezeBalanceV2: %+v \n", err)
		return
	}
	fmt.Println("rawTx:", rawTx.String())

	signTransaction(rawTx.Transaction)
	txBroad, err := conn.Broadcast(rawTx.Transaction)
	if err != nil {
		fmt.Printf("fail to Broadcast: %+v \n", err)
		return
	}
	fmt.Println("txBroad:", txBroad.String())
}

func signTransaction(tx *core.Transaction) (*core.Transaction, error) {
	// Decode the private key from hex
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Get the raw transaction hash
	rawData, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("failed to serialize transaction: %w", err)
	}

	hash := sha256.Sum256(rawData)

	// Sign the hash with the private key
	signature, err := crypto.Sign(hash[:], privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}

	// Add the signature to the transaction
	tx.Signature = append(tx.Signature, signature)
	return tx, nil
}

func voteTransaction() {
	conn := client.NewGrpcClient(shastaDomain)
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		fmt.Printf("fail to NewGrpcClient: %+v \n", conn)
		return
	}
	defer conn.Stop()

	// Define the votes
	votes := []*core.VoteWitnessContract_Vote{
		{
			VoteAddress: decodeAddress("TExYAkRVJsRdbAPfdFtGKPE7ZxCDopq8dQ"),
			VoteCount:   34, // Number of votes for this SR
		},
		{
			VoteAddress: decodeAddress("TVDuR9wjfhVDZYPB1YsTu5f5QVLxE2JqLS"),
			VoteCount:   34, // Number of votes for this SR
		},
		{
			VoteAddress: decodeAddress("TDcHeAMHPYfJKsknLCnV89K9wu8dqdD3rm"),
			VoteCount:   33, // Number of votes for this SR
		},
	}

	// Create the vote transaction
	voteContract := &core.VoteWitnessContract{
		OwnerAddress: decodeAddress(ownAddress),
		Votes:        votes,
	}
	rawTx, err := conn.Client.VoteWitnessAccount2(context.Background(), voteContract)
	if err != nil {
		log.Fatalf("Failed to create vote transaction: %v", err)
	}
	fmt.Printf("Raw vote transaction created: %s\n", rawTx.String())

	// Sign the transaction with the private key
	signedTx, err := signTransaction(rawTx.Transaction)
	if err != nil {
		log.Fatalf("Failed to sign vote transaction: %v", err)
	}
	fmt.Printf("Signed vote transaction: %s\n", signedTx.String())

	// Broadcast the signed transaction
	result, err := conn.Broadcast(signedTx)
	if err != nil {
		log.Fatalf("Failed to broadcast vote transaction: %v", err)
	}
	fmt.Printf("Vote transaction broadcast result: %s\n", result.String())

}

// decodeAddress converts a base58 Tron address to a hex format
func decodeAddress(base58Addr string) []byte {
	decoded := base58.Decode(base58Addr)
	if len(decoded) == 0 {
		log.Fatalf("Failed to decode base58 address: %s", base58Addr)
	}

	// The last 4 bytes are the checksum, so we only take the first 21 bytes
	return decoded[:len(decoded)-4]
}

func encodeAddress(decodeAddr []byte) string {
	if len(decodeAddr) != 21 {
		panic("Invalid raw address length. Expected 21 bytes.")
	}

	encoded := base58.CheckEncode(decodeAddr[1:], decodeAddr[0])
	if len(encoded) == 0 {
		log.Fatalf("Failed to decode base58 address: %s", decodeAddr)
	}

	// The last 4 bytes are the checksum, so we only take the first 21 bytes
	return encoded
}

func getVotePower() {
	// Connect to the Tron network
	conn := client.NewGrpcClient(shastaDomain)
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Tron network: %v", err)
	}
	defer conn.Stop()

	// Query account information
	account, err := conn.GetAccount(ownAddress)
	if err != nil {
		log.Fatalf("Failed to get account information: %v", err)
	}

	// Calculate total voting power
	totalVotePower := int64(0)
	for _, v := range account.FrozenV2 {
		totalVotePower += v.Amount
	}
	totalVotePower = int64(float64(totalVotePower) / 1_000_000)

	totalVotedPower := int64(0)
	for _, v := range account.Votes {
		totalVotedPower += v.VoteCount
	}

	fmt.Printf("Total Vote Power (in TRX): %d\n", totalVotePower)
	fmt.Printf("Total Voted Power (in TRX): %d\n", totalVotedPower)
}

func getSpAddress() {
	// Connect to the Tron network
	conn := client.NewGrpcClient(shastaDomain)
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Tron network: %v", err)
	}
	defer conn.Stop()

	// Get SR list
	witnessList, err := conn.ListWitnesses()
	if err != nil {
		log.Fatalf("Failed to get SR list: %v", err)
	}

	// Sort witnesses by VoteCount in descending order
	sort.Slice(witnessList.Witnesses, func(i, j int) bool {
		return witnessList.Witnesses[i].VoteCount > witnessList.Witnesses[j].VoteCount
	})

	// Print SR addresses
	fmt.Println("List of Super Representatives:")
	for _, witness := range witnessList.Witnesses {
		fmt.Printf("Address: %s, URL: %s, Last Round Votes: %d, Total Produced: %d \n\n", encodeAddress(witness.Address), witness.Url, witness.VoteCount, witness.TotalProduced)
	}
}

func getClaimableReward() {
	// Initialize the gRPC client
	conn := client.NewGrpcClient(shastaDomain)
	err := conn.Start(grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Failed to start gRPC client: %+v\n", err)
		return
	}
	defer conn.Stop()

	// Get Claimable Rewards
	claimableRewards, err := conn.GetRewardsInfo(ownAddress)
	if err != nil {
		fmt.Printf("Error fetching rewards: %+v\n", err)
		return
	}

	// Display Rewards
	fmt.Printf("Claimable Rewards for address %s: %.6f TRX\n", ownAddress, float64(claimableRewards)/1_000_000)

}
