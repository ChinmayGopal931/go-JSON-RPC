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

func SwapPermit(c *gin.Context) {
	// Hardcoded values (replace with actual values from your environment)
	currency0 := common.HexToAddress("0xAA292E8611aDF267e563f334Ee42320aC96D0463")
	currency1 := common.HexToAddress("0xf953b3A269d80e3eB0F2947630Da976B896A8C5b")

	zeroForOne := true
	amountSpecified, _ := new(big.Int).SetString("10000000", 10) // 1 ether

	sqrtPriceLimitX96, _ := new(big.Int).SetString("4295128740", 10)
	// userAddress := common.HexToAddress("0x328809Bc894f92807417D2dAD6b7C998c1aFdac6")

	userAddress, alicePrivKey := utils.MakeAddrAndKey("alice")
	fmt.Printf("Alice's address: %s\n", userAddress.Hex())
	fmt.Printf("Alice's private key: 0x%x\n", crypto.FromECDSA(alicePrivKey))

	log.Println("alice userAddress", userAddress, alicePrivKey)

	// Create the pool key
	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	// Prepare swap parameters
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

	log.Printf("PoolKey: currency0=%s, currency1=%s, fee=%d, tickSpacing=%d, hooks=%s",
		poolKey.Currency0.Hex(), poolKey.Currency1.Hex(), poolKey.Fee, poolKey.TickSpacing, poolKey.Hooks.Hex())
	log.Printf("SwapParams: zeroForOne=%v, amountSpecified=%s, sqrtPriceLimitX96=%s",
		swapParams.ZeroForOne, swapParams.AmountSpecified.String(), swapParams.SqrtPriceLimitX96.String())

	// Prepare permit data
	deadline := big.NewInt(time.Now().Unix() + 3600) // 1 hour from now
	value := new(big.Int).Mul(amountSpecified, big.NewInt(11))
	value = value.Div(value, big.NewInt(10)) // Increase by 10% to account for fees and slippage

	log.Printf("Token Address (currency0): %s", currency0.Hex())
	log.Printf("Spender Address (SwapRouterAddress): %s", ethereum.SwapRouterAddress.Hex())
	log.Printf("User Address: %s", userAddress.Hex())
	log.Printf("Value: %s", value.String())
	log.Printf("Deadline: %s", deadline.String())

	// Generate permit signature
	v, r, s, err := utils.GeneratePermitSignature(currency0, userAddress, ethereum.SwapRouterAddress, value, deadline, alicePrivKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to generate permit signature: %v", err)})
		return
	}

	// Pack the data for the swapWithPermit function call
	data, err := ethereum.SwapRouterABI.Pack("swapWithPermit",
		userAddress,
		poolKey,
		swapParams,
		testSettings,
		[]byte{}, // hookData
		deadline,
		v,
		r,
		s,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error packing data: %v", err)})
		return
	}
	chainID, _ := ethereum.Client.ChainID(context.Background())

	auth, _ := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)

	balance0Before, err := utils.GetBalance(currency0, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency0 before swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1Before, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 before swap: %v", err)
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

	tx := types.NewTransaction(nonce, ethereum.SwapRouterAddress, big.NewInt(0), 1000000, gasPrice, data)

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
		log.Printf("Error getting balance of currency0 after swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	balance1After, err := utils.GetBalance(currency1, userAddress)
	if err != nil {
		log.Printf("Error getting balance of currency1 after swap: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	c.JSON(200, gin.H{
		"txHash":         signedTx.Hash().Hex(),
		"message":        "Swap with permit initiated successfully",
		"balancesBefore": gin.H{"currency0": balance0Before.String(), "currency1": balance1Before.String()},
		"balancesAfter":  gin.H{"currency0": balance0After.String(), "currency1": balance1After.String()},
		"deltaBalances":  gin.H{"currency0": delta0.String(), "currency1": delta1.String()},
	})
}
