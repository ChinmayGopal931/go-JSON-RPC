package handlers

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"
	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/pkg/utils"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

type AddLiquidityRequest struct {
	Currency0   string `json:"currency0" binding:"required"`
	Currency1   string `json:"currency1" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
	UserAddress string `json:"userAddress" binding:"required"`
	PrivateKey  string `json:"privateKey" binding:"required"`
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
	amount, success := new(big.Int).SetString(req.Amount, 10)
	if !success {
		c.JSON(400, gin.H{"error": "Invalid amount value"})
		return
	}
	userAddress := common.HexToAddress(req.UserAddress)
	// Parse private key
	privateKey, err := crypto.HexToECDSA(req.PrivateKey)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid private key: " + err.Error()})
		return
	}

	// Hardcoded tick range
	minTick := big.NewInt(-887220)
	maxTick := big.NewInt(887220)

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
		LiquidityDelta: amount,
		Salt:           [32]byte{},
	}

	// Prepare permit data
	deadline := big.NewInt(time.Now().Unix() + 3600) // 1 hour from now
	value := new(big.Int).Mul(amount, big.NewInt(10))

	// Generate permit signatures for both tokens
	v0, r0, s0, err := utils.GeneratePermitSignature(currency0, userAddress, ethereum.LPRouterAddress, value, deadline, privateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error generating permit signature for currency0: " + err.Error()})
		return
	}

	v1, r1, s1, err := utils.GeneratePermitSignature(currency1, userAddress, ethereum.LPRouterAddress, value, deadline, privateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error generating permit signature for currency1: " + err.Error()})
		return
	}

	// Pack the data for the modifyLiquidityWithPermit function call
	data, err := ethereum.LPRouterABI.Pack("modifyLiquidityWithPermit",
		userAddress,
		poolKey,
		params,
		[]byte{}, // hookData
		false,    // settleUsingBurn
		false,    // takeClaims
		deadline,
		v0, r0, s0,
		v1, r1, s1,
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
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
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

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
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
