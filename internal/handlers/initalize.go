package handlers

import (
	"context"
	"fmt"
	"math/big"

	"uniswap-v4-rpc/internal/ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
)

func Initialize(c *gin.Context) {
	var req struct {
		Token0      common.Address `json:"token0"`
		Token1      common.Address `json:"token1"`
		TickSpacing int64          `json:"tickSpacing"`
		Hook        common.Address `json:"hook"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	auth, err := createTransactor()
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to create transactor: %v", err)})
		return
	}

	// Constants
	sqrtPrice1To1, _ := new(big.Int).SetString("79228162514264337593543950336", 10)

	// Create PoolKey
	currency0 := common.HexToAddress("0xAA292E8611aDF267e563f334Ee42320aC96D0463")
	currency1 := common.HexToAddress("0xf953b3A269d80e3eB0F2947630Da976B896A8C5b")

	poolKey := createPoolKey(currency0, currency1, ethereum.HookAddress)

	initData, err := ethereum.ManagerABI.Pack("initialize", poolKey, sqrtPrice1To1, []byte{})
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to pack initialize data: %v", err)})
		return
	}

	tx := types.NewTransaction(auth.Nonce.Uint64(), ethereum.ManagerAddress, big.NewInt(0), 500000, auth.GasPrice, initData)
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to sign initialize transaction: %v", err)})
		return
	}

	err = ethereum.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to send initialize transaction: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"initializeTxHash": signedTx.Hash().Hex(),
		"status":           "Pool initialized successfully",
	})
}

func createTransactor() (*bind.TransactOpts, error) {
	chainID, err := ethereum.Client.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(ethereum.PrivateKey, chainID)
	if err != nil {
		return nil, err
	}

	nonce, err := ethereum.Client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, err
	}
	auth.Nonce = big.NewInt(int64(nonce))

	gasPrice, err := ethereum.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}
	auth.GasPrice = gasPrice

	return auth, nil
}

func createPoolKey(token0, token1 common.Address, hook common.Address) struct {
	Currency0   common.Address
	Currency1   common.Address
	Fee         *big.Int
	TickSpacing *big.Int
	Hooks       common.Address
} {
	return struct {
		Currency0   common.Address
		Currency1   common.Address
		Fee         *big.Int
		TickSpacing *big.Int
		Hooks       common.Address
	}{
		Currency0:   token0,
		Currency1:   token1,
		Fee:         big.NewInt(3000), // harcoded fee, need to adjust as needed
		TickSpacing: big.NewInt(60),   // harcoded tickspacing, need to adjust as needed
		Hooks:       hook,
	}
}
