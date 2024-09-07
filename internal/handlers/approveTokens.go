package handlers

import (
	"fmt"
	"log"

	"uniswap-v4-rpc/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// ApproveTokens handles the approval of both hardcoded tokens for the SwapRouter and LPRouter
func ApproveTokens(c *gin.Context) {

	currency0 := common.HexToAddress("0xAA292E8611aDF267e563f334Ee42320aC96D0463")
	currency1 := common.HexToAddress("0xf953b3A269d80e3eB0F2947630Da976B896A8C5b")

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
