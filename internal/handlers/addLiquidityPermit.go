package handlers

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
)

// AddLiquidityRequest represents the request structure for adding liquidity with permits
type AddLiquidityRequest struct {
	Currency0   string   `json:"currency0" binding:"required"`
	Currency1   string   `json:"currency1" binding:"required"`
	Amount      string   `json:"amount" binding:"required"`
	UserAddress string   `json:"userAddress" binding:"required"`
	Deadline    *big.Int `json:"deadline" binding:"required"`
	// Signature for currency0
	V0 uint8  `json:"v0" binding:"required"`
	R0 string `json:"r0" binding:"required"`
	S0 string `json:"s0" binding:"required"`
	// Signature for currency1
	V1 uint8  `json:"v1" binding:"required"`
	R1 string `json:"r1" binding:"required"`
	S1 string `json:"s1" binding:"required"`
}

// AddLiquidityPermit handles the addition of liquidity using signed permits
// This allows users to add liquidity without needing to pre-approve tokens
func AddLiquidityPermit(c *gin.Context) {
	// Parse and validate request body
	var req AddLiquidityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Convert addresses and amount
	currency0 := common.HexToAddress(req.Currency0)
	currency1 := common.HexToAddress(req.Currency1)
	userAddress := common.HexToAddress(req.UserAddress)
	amount, success := new(big.Int).SetString(req.Amount, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid amount value"})
		return
	}

	// Convert signature components to the required format
	var r0, s0, r1, s1 [32]byte
	copy(r0[:], common.FromHex(req.R0))
	copy(s0[:], common.FromHex(req.S0))
	copy(r1[:], common.FromHex(req.R1))
	copy(s1[:], common.FromHex(req.S1))

	// Log initial parameters
	log.Printf("Adding liquidity for user: %s", userAddress.Hex())
	log.Printf("Currency0: %s, Currency1: %s", currency0.Hex(), currency1.Hex())
	log.Printf("Amount: %s", amount.String())

	// Define liquidity parameters
	// Using full range for now (-887220 to 887220)
	params := struct {
		TickLower      *big.Int
		TickUpper      *big.Int
		LiquidityDelta *big.Int
		Salt           [32]byte
	}{
		TickLower:      big.NewInt(-887220),
		TickUpper:      big.NewInt(887220),
		LiquidityDelta: amount,
		Salt:           [32]byte{},
	}

	// Create the pool key with standard parameters
	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	// Log pool key details for debugging
	log.Printf("Pool Key - Currency0: %s, Currency1: %s, Hooks: %s",
		poolKey.Currency0.Hex(),
		poolKey.Currency1.Hex(),
		poolKey.Hooks.Hex(),
	)

	// Get initial balances for comparison
	balance0Before, err := utils.GetBalance(currency0, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency0 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get initial balance"})
		return
	}
	balance1Before, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get initial balance"})
		return
	}

	// Pack the transaction data
	data, err := ethereum.LPRouterABI.Pack("modifyLiquidityWithPermit",
		userAddress,
		poolKey,
		params,
		[]byte{}, // hookData
		false,    // settleUsingBurn
		false,    // takeClaims
		req.Deadline,
		req.V0, r0, s0, // Currency0 signature
		req.V1, r1, s1, // Currency1 signature
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error packing transaction data: " + err.Error()})
		return
	}

	// Prepare transaction parameters
	chainID, err := ethereum.Client.ChainID(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to get chain ID: " + err.Error()})
		return
	}

	// Create transactor using server's private key
	auth, err := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create transactor: " + err.Error()})
		return
	}

	// Get the current nonce and gas price
	nonce, err := ethereum.Client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get nonce: %v", err)})
		return
	}

	gasPrice, err := ethereum.Client.SuggestGasPrice(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to get gas price: %v", err)})
		return
	}

	// Create and sign transaction
	tx := types.NewTransaction(nonce, ethereum.LPRouterAddress, big.NewInt(0), 1000000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ethereum.PrivateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to sign transaction: %v", err)})
		return
	}

	// Send transaction
	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to send transaction: %v", err)})
		return
	}

	// Get final balances
	balance0After, err := utils.GetBalance(currency0, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency0 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get final balance"})
		return
	}
	balance1After, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get final balance"})
		return
	}

	// Calculate balance changes
	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	// Return success response with details
	c.JSON(200, gin.H{
		"txHash":  signedTx.Hash().Hex(),
		"message": "Liquidity added successfully with permits",
		"details": gin.H{
			"balancesBefore": gin.H{
				"currency0": balance0Before.String(),
				"currency1": balance1Before.String(),
			},
			"balancesAfter": gin.H{
				"currency0": balance0After.String(),
				"currency1": balance1After.String(),
			},
			"deltaBalances": gin.H{
				"currency0": delta0.String(),
				"currency1": delta1.String(),
			},
		},
		"poolKey": gin.H{
			"currency0": poolKey.Currency0.Hex(),
			"currency1": poolKey.Currency1.Hex(),
			"fee":       poolKey.Fee,
			"hooks":     poolKey.Hooks.Hex(),
		},
	})
}
