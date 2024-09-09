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

type AddLiquidityRequest struct {
	Currency0       string `json:"currency0"`
	Currency1       string `json:"currency1"`
	MinTick         string `json:"minTick"`
	MaxTick         string `json:"maxTick"`
	LiquidityAmount string `json:"liquidityAmount"`
	UserAddress     string `json:"userAddress"`
	Deadline        string `json:"deadline"`
	V0              uint8  `json:"v0"`
	R0              string `json:"r0"`
	S0              string `json:"s0"`
	V1              uint8  `json:"v1"`
	R1              string `json:"r1"`
	S1              string `json:"s1"`
}

func AddLiquidityPermit(c *gin.Context) {
	var req AddLiquidityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Convert string inputs to appropriate types
	currency0 := common.HexToAddress(req.Currency0)
	currency1 := common.HexToAddress(req.Currency1)
	minTick, success := new(big.Int).SetString(req.MinTick, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid minTick value"})
		return
	}
	maxTick, success := new(big.Int).SetString(req.MaxTick, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid maxTick value"})
		return
	}
	liquidityAmount, success := new(big.Int).SetString(req.LiquidityAmount, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid liquidityAmount value"})
		return
	}
	userAddress := common.HexToAddress(req.UserAddress)
	deadline, success := new(big.Int).SetString(req.Deadline, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid deadline value"})
		return
	}

	// Create the pool key
	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	// Prepare modifyLiquidity parameters
	params := struct {
		TickLower      *big.Int
		TickUpper      *big.Int
		LiquidityDelta *big.Int
		Salt           [32]byte
	}{
		TickLower:      minTick,
		TickUpper:      maxTick,
		LiquidityDelta: liquidityAmount,
		Salt:           [32]byte{},
	}

	// Convert signature components
	r0 := common.HexToHash(req.R0)
	s0 := common.HexToHash(req.S0)
	r1 := common.HexToHash(req.R1)
	s1 := common.HexToHash(req.S1)

	// Pack the data for the modifyLiquidityWithPermit function call
	data, err := ethereum.LPRouterABI.Pack("modifyLiquidityWithPermit",
		userAddress,
		poolKey,
		params,
		[]byte{}, // hookData
		false,    // settleUsingBurn
		false,    // takeClaims
		deadline,
		req.V0, r0, s0,
		req.V1, r1, s1,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error packing data: " + err.Error()})
		return
	}

	chainID, err := ethereum.Client.ChainID(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": "Error getting chain ID: " + err.Error()})
		return
	}
	auth, err := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error creating transactor: " + err.Error()})
		return
	}

	balance0Before, err := utils.GetBalance(currency0, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency0 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1Before, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Create and send the transaction
	nonce, err := ethereum.Client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error fetching nonce: %v", err)})
		return
	}

	gasPrice, err := ethereum.Client.SuggestGasPrice(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error fetching gas price: %v", err)})
		return
	}

	tx := types.NewTransaction(nonce, ethereum.LPRouterAddress, big.NewInt(0), 1000000, gasPrice, data)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ethereum.PrivateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error signing transaction: %v", err)})
		return
	}

	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error sending transaction: %v", err)})
		return
	}

	balance0After, err := utils.GetBalance(currency0, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency0 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1After, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	c.JSON(200, gin.H{
		"txHash":         signedTx.Hash().Hex(),
		"message":        "Add liquidity with permit initiated successfully",
		"balancesBefore": gin.H{"currency0": balance0Before.String(), "currency1": balance1Before.String()},
		"balancesAfter":  gin.H{"currency0": balance0After.String(), "currency1": balance1After.String()},
		"deltaBalances":  gin.H{"currency0": delta0.String(), "currency1": delta1.String()},
	})
}
