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

func AddLiquidity(c *gin.Context) {
	// Hardcoded values (replace with actual values from your environment)
	currency0 := common.HexToAddress("0xAA292E8611aDF267e563f334Ee42320aC96D0463")
	currency1 := common.HexToAddress("0xf953b3A269d80e3eB0F2947630Da976B896A8C5b")

	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)

	minTick := big.NewInt(-887220)
	maxTick := big.NewInt(887220)
	liquidityAmount, _ := new(big.Int).SetString("100000000000000000000", 10) // 100 ether

	log.Printf("Adding liquidity with the following parameters:")
	log.Printf("Currency0: %s", currency0.Hex())
	log.Printf("Currency1: %s", currency1.Hex())
	log.Printf("Fee: %s", fee.String())
	log.Printf("TickSpacing: %s", tickSpacing.String())
	log.Printf("LiquidityAmount: %s", liquidityAmount.String())

	auth, err := createTransactor()
	if err != nil {
		log.Printf("Failed to create transactor: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create transactor: %v", err)})
		return
	}

	networkID, err := ethereum.Client.NetworkID(context.Background())
	if err != nil {
		log.Printf("Failed to get network ID: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get network ID"})
		return
	}
	log.Printf("Connected to network with ID: %s", networkID.String())

	log.Printf("Transactor created with address: %s", auth.From.Hex())

	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	if err := utils.CheckContractDeployment(currency0); err != nil {
		log.Printf("Error with currency0 contract: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Currency0 contract issue: %v", err)})
		return
	}
	if err := utils.CheckContractDeployment(currency1); err != nil {
		log.Printf("Error with currency1 contract: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Currency1 contract issue: %v", err)})
		return
	}

	// Check balances before adding liquidity
	balance0Before, err := utils.GetBalance(currency0, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency0 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error getting balance of currency0: %v", err)})
		return
	}
	log.Printf("Balance of currency0 before: %s", balance0Before.String())

	balance1Before, err := utils.GetBalance(currency1, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency1 before adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error getting balance of currency1: %v", err)})
		return
	}
	log.Printf("Balance of currency1 before: %s", balance1Before.String())

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

	data, err := ethereum.LPRouterABI.Pack("modifyLiquidity", poolKey, params, []byte{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to pack data: %v", err)})
		return
	}

	tx := types.NewTransaction(auth.Nonce.Uint64(), ethereum.LPRouterAddress, big.NewInt(0), 500000, auth.GasPrice, data)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to sign transaction: %v", err)})
		return
	}

	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to send transaction: %v", err)})
		return
	}

	// Check balances after adding liquidity
	balance0After, err := utils.GetBalance(currency0, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency0 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1After, err := utils.GetBalance(currency1, auth.From)
	if err != nil {
		log.Printf("Error getting balance of currency1 after adding liquidity: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	c.JSON(200, gin.H{
		"status":         "Liquidity added successfully",
		"txHash":         signedTx.Hash().Hex(),
		"balancesBefore": gin.H{"currency0": balance0Before.String(), "currency1": balance1Before.String()},
		"balancesAfter":  gin.H{"currency0": balance0After.String(), "currency1": balance1After.String()},
		"deltaBalances":  gin.H{"currency0": delta0.String(), "currency1": delta1.String()},
		"params": gin.H{
			"currency0":       currency0.Hex(),
			"currency1":       currency1.Hex(),
			"fee":             fee.String(),
			"tickSpacing":     tickSpacing.String(),
			"liquidityAmount": liquidityAmount.String(),
		},
	})
}
