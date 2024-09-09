package handlers

import (
	"fmt"
	"log"

	"uniswap-v4-rpc/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// ApproveTokens handles the approval of both tokens for the SwapRouter and LPRouter
func ApproveTokens(c *gin.Context) {
	var req struct {
		Currency0 string `json:"currency0" binding:"required"`
		Currency1 string `json:"currency1" binding:"required"`
	}

	log.Println("eeeeee", req.Currency0, req.Currency1)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	currency0 := common.HexToAddress(req.Currency0)
	currency1 := common.HexToAddress(req.Currency1)
	log.Println(currency0, currency1)

	// Create transactor
	auth, err := createTransactor()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create transactor: %v", err)})
		return
	}

	// Use the ApproveTokens function from utils
	err = utils.ApproveTokens(auth, currency0, currency1)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to approve tokens: %v", err)})
		return
	}

	// Get balances after approval
	results := make(map[string]interface{})
	for _, currency := range []common.Address{currency0, currency1} {
		balance, err := utils.GetBalance(currency, auth.From)
		if err != nil {
			log.Printf("Error getting balance for %s: %v", currency.Hex(), err)
			results[currency.Hex()] = fmt.Sprintf("Failed to get balance: %v", err)
		} else {
			results[currency.Hex()] = gin.H{
				"message": "Token approved successfully",
				"balance": balance.String(),
			}
		}
	}

	c.JSON(200, results)
}
