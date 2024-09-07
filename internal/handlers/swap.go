package handlers

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/pkg/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
)

func Swap(c *gin.Context) {
	// Hardcoded values (consider making these configurable or part of the request)
	currency0 := common.HexToAddress("0xAA292E8611aDF267e563f334Ee42320aC96D0463")
	currency1 := common.HexToAddress("0xf953b3A269d80e3eB0F2947630Da976B896A8C5b")

	zeroForOne := true
	amountSpecified, _ := new(big.Int).SetString("10000000", 10) // 1 ether
	sqrtPriceLimitX96, _ := new(big.Int).SetString("4295128740", 10)

	auth, err := createTransactor()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create transactor: %v", err)})
		return
	}

	balance0Before, err := utils.GetBalance(currency0, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency0 before swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1Before, err := utils.GetBalance(currency1, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency1 before swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	swapParams := struct {
		ZeroForOne        bool
		AmountSpecified   *big.Int
		SqrtPriceLimitX96 *big.Int
	}{
		ZeroForOne:        zeroForOne,
		AmountSpecified:   amountSpecified,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}

	testSettings := struct {
		TakeClaims      bool
		SettleUsingBurn bool
	}{
		TakeClaims:      false,
		SettleUsingBurn: false,
	}

	data, err := ethereum.SwapRouterABI.Pack("swap", poolKey, swapParams, testSettings, []byte{})
	if err != nil {
		log.Printf("Error packing data: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	tx := types.NewTransaction(auth.Nonce.Uint64(), ethereum.SwapRouterAddress, big.NewInt(0), 1000000, auth.GasPrice, data)

	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		log.Printf("Error signing transaction: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Printf("Error sending transaction: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	balance0After, err := utils.GetBalance(currency0, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency0 after swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1After, err := utils.GetBalance(currency1, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency1 after swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	c.JSON(200, gin.H{
		"txHash":         signedTx.Hash().Hex(),
		"balancesBefore": gin.H{"currency0": balance0Before.String(), "currency1": balance1Before.String()},
		"balancesAfter":  gin.H{"currency0": balance0After.String(), "currency1": balance1After.String()},
		"deltaBalances":  gin.H{"currency0": delta0.String(), "currency1": delta1.String()},
	})
}
