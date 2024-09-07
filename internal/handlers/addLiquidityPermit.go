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

func AddLiquidityPermit(c *gin.Context) {
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

	userAddress, userPrivKey := utils.MakeAddrAndKey("alice")
	fmt.Printf("User's address: %s\n", userAddress.Hex())
	fmt.Printf("User's private key: 0x%x\n", crypto.FromECDSA(userPrivKey))

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

	// Prepare permit data
	deadline := big.NewInt(time.Now().Unix() + 3600) // 1 hour from now
	value := new(big.Int).Set(liquidityAmount)       // Use liquidityAmount as the value for both tokens

	log.Printf("Token0 Address: %s", currency0.Hex())
	log.Printf("Token1 Address: %s", currency1.Hex())
	log.Printf("LPRouter Address: %s", ethereum.LPRouterAddress.Hex())
	log.Printf("User Address: %s", userAddress.Hex())
	log.Printf("Value: %s", value.String())
	log.Printf("Deadline: %s", deadline.String())

	// Generate permit signatures for both tokens
	v0, r0, s0, err := utils.GeneratePermitSignature(currency0, userAddress, ethereum.LPRouterAddress, value, deadline, userPrivKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate permit signature for token0: %v", err)})
		return
	}

	v1, r1, s1, err := utils.GeneratePermitSignature(currency1, userAddress, ethereum.LPRouterAddress, value, deadline, userPrivKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate permit signature for token1: %v", err)})
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
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error packing data: %v", err)})
		return
	}

	chainID, _ := ethereum.Client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)

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
