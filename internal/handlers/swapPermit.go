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

// SwapRequest defines the structure for a swap request with permit signature
type SwapRequest struct {
	Currency0   string   `json:"currency0" binding:"required"`   // Address of the input token
	Currency1   string   `json:"currency1" binding:"required"`   // Address of the output token
	Amount      string   `json:"amount" binding:"required"`      // Amount to swap in base units (wei)
	ZeroForOne  bool     `json:"zeroForOne"`                     // Direction of swap (true for currency0 to currency1)
	UserAddress string   `json:"userAddress" binding:"required"` // Address of the user initiating the swap
	SignatureV  uint8    `json:"v" binding:"required"`           // V component of the EIP-712 signature
	SignatureR  string   `json:"r" binding:"required"`           // R component of the EIP-712 signature
	SignatureS  string   `json:"s" binding:"required"`           // S component of the EIP-712 signature
	Deadline    *big.Int `json:"deadline" binding:"required"`    // Timestamp after which the permit becomes invalid
}

// SwapPermit handles token swaps using signed permits
// This function allows users to swap tokens without pre-approving them,
// by using EIP-712 signatures (gasless approvals)
func SwapPermit(c *gin.Context) {
	// Parse and validate the incoming JSON request
	var req SwapRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Convert string addresses to Ethereum addresses
	currency0 := common.HexToAddress(req.Currency0)
	currency1 := common.HexToAddress(req.Currency1)
	userAddress := common.HexToAddress(req.UserAddress)

	// Parse the amount string to big.Int
	amountSpecified, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		c.JSON(400, gin.H{"error": "Invalid amount"})
		return
	}

	// Convert hex signature strings to bytes32 format
	var r, s [32]byte
	rBytes := common.FromHex(req.SignatureR)
	sBytes := common.FromHex(req.SignatureS)
	copy(r[:], rBytes)
	copy(s[:], sBytes)

	// Set swap direction and price limit
	zeroForOne := req.ZeroForOne
	// Default price limit for swaps
	sqrtPriceLimitX96, _ := new(big.Int).SetString("4295128740", 10)

	// Log user address for debugging
	fmt.Printf("User's address: %s\n", userAddress.Hex())

	// Create pool key for the token pair
	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	// Define swap parameters structure
	swapParams := struct {
		ZeroForOne        bool
		AmountSpecified   *big.Int
		SqrtPriceLimitX96 *big.Int
	}{
		ZeroForOne:        zeroForOne,
		AmountSpecified:   amountSpecified,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}

	// Define test settings for the swap
	testSettings := struct {
		TakeClaims      bool
		SettleUsingBurn bool
	}{
		TakeClaims:      false,
		SettleUsingBurn: false,
	}

	// Log pool and swap details for debugging
	log.Printf("PoolKey: currency0=%s, currency1=%s, fee=%d, tickSpacing=%d, hooks=%s",
		poolKey.Currency0.Hex(), poolKey.Currency1.Hex(), poolKey.Fee, poolKey.TickSpacing, poolKey.Hooks.Hex())
	log.Printf("SwapParams: zeroForOne=%v, amountSpecified=%s, sqrtPriceLimitX96=%s",
		swapParams.ZeroForOne, swapParams.AmountSpecified.String(), swapParams.SqrtPriceLimitX96.String())

	// Calculate value including buffer for fees and slippage (10% buffer)
	value := new(big.Int).Mul(amountSpecified, big.NewInt(11))
	value = value.Div(value, big.NewInt(10))

	// Log transaction details
	log.Printf("Token Address (currency0): %s", currency0.Hex())
	log.Printf("Spender Address (SwapRouterAddress): %s", ethereum.SwapRouterAddress.Hex())
	log.Printf("User Address: %s", userAddress.Hex())
	log.Printf("Value: %s", value.String())
	log.Printf("Deadline: %s", req.Deadline.String())
	log.Printf("Signature (v,r,s): %d, 0x%x, 0x%x", req.SignatureV, r, s)

	// Pack the transaction data for the swapWithPermit function
	data, err := ethereum.SwapRouterABI.Pack("swapWithPermit",
		userAddress,
		poolKey,
		swapParams,
		testSettings,
		[]byte{}, // hookData (empty for standard swaps)
		req.Deadline,
		req.SignatureV,
		r,
		s,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error packing data: %v", err)})
		return
	}

	// Get chain ID and create transactor
	chainID, _ := ethereum.Client.ChainID(context.Background())
	auth, _ := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)

	// Get initial balances for comparison
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

	// Get current nonce for transaction
	nonce, err := ethereum.Client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error fetching nonce: %v", err)})
		return
	}

	// Get current gas price
	gasPrice, err := ethereum.Client.SuggestGasPrice(context.Background())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error fetching gas price: %v", err)})
		return
	}

	// Create and sign transaction
	tx := types.NewTransaction(nonce, ethereum.SwapRouterAddress, big.NewInt(0), 1000000, gasPrice, data)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), ethereum.PrivateKey)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error signing transaction: %v", err)})
		return
	}

	// Send transaction to the network
	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Error sending transaction: %v", err)})
		return
	}

	// Get final balances after swap
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

	// Calculate balance changes
	delta0 := new(big.Int).Sub(balance0After, balance0Before)
	delta1 := new(big.Int).Sub(balance1After, balance1Before)

	// Return success response with transaction details
	c.JSON(200, gin.H{
		"txHash":  signedTx.Hash().Hex(),
		"message": "Swap with permit initiated successfully",
		"balances": gin.H{
			"before": gin.H{
				"currency0": balance0Before.String(),
				"currency1": balance1Before.String(),
			},
			"after": gin.H{
				"currency0": balance0After.String(),
				"currency1": balance1After.String(),
			},
			"delta": gin.H{
				"currency0": delta0.String(),
				"currency1": delta1.String(),
			},
		},
	})
}
