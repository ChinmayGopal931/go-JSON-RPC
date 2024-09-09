package integration

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/pkg/utils"
)

func TestAddressesInitialized(t *testing.T) {
	addresses := []struct {
		name    string
		address common.Address
	}{
		{"SwapRouter", ethereum.SwapRouterAddress},
		{"LPRouter", ethereum.LPRouterAddress},
		{"Manager", ethereum.ManagerAddress},
		{"Hook", ethereum.HookAddress},
		{"token1", ethereum.Token0_address},
		{"token1", ethereum.Token1_address},
	}

	for _, addr := range addresses {
		t.Run(addr.name, func(t *testing.T) {
			// Check if address is not zero
			assert.NotEqual(t, common.Address{}, addr.address, "Address should not be zero")

			// Check if contract exists at the address
			err := utils.CheckContractDeployment(addr.address)
			assert.NoError(t, err, "Contract should exist at the address")
		})
	}
}

func TestABIsInitialized(t *testing.T) {
	abis := []struct {
		name string
		abi  abi.ABI
	}{
		{"SwapRouter", ethereum.SwapRouterABI},
		{"LPRouter", ethereum.LPRouterABI},
		{"Manager", ethereum.ManagerABI},
	}

	for _, a := range abis {
		t.Run(a.name, func(t *testing.T) {
			assert.NotEmpty(t, a.abi.Methods, "ABI should have methods")
		})
	}
}
