package handler

import (
	"errors"
	"fmt"
	"net/http"

	// base58 "github.com/btcsuite/btcd/btcutil/base58"

	"github.com/gin-gonic/gin"
	cf "github.com/kahsengphoon/Tron-Votes/config"
	"github.com/urfave/cli/v2"
)

const (
	shastaDomain = "https://api.shasta.trongrid.io"
	ownAddress   = "TKjxS5iEAzyJHY9g7ZJR8XwTshDUEeH379"
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

	if err := r.Run(fmt.Sprintf(":%v", cf.Enviroment().AppServerPort)); err != nil {
		return err
	}

	return nil
}

// func freezeBalance() {
// 	client := resty.New()

// 	// Convert Base58 address to Hex
// 	ownerAddressHex, err := base58ToHex(ownAddress)
// 	if err != nil {
// 		log.Fatalf("Address conversion error: %v", err)
// 	}

// 	// Define the freeze request
// 	body := map[string]interface{}{
// 		"owner_address":   ownerAddressHex, // Hex address
// 		"frozen_balance":  1000000,         // Amount in SUN (1 TRX = 1,000,000 SUN)
// 		"frozen_duration": 1,               // Freeze duration in days
// 		"resource":        "ENERGY",        // Options: "ENERGY" or "BANDWIDTH"
// 	}

// 	// Send the request
// 	resp, err := client.R().
// 		SetHeader("Content-Type", "application/json").
// 		SetBody(body).
// 		Post(shastaDomain + "/wallet/freezebalance")

// 	if err != nil {
// 		log.Fatalf("Error freezing balance: %v", err)
// 	}

// 	fmt.Println("Response:", resp.String())
// }
