// test/integration/liquidity_test.go

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"uniswap-v4-rpc/internal/ethereum"

	"github.com/stretchr/testify/assert"
)

func TestAddLiquidity(t *testing.T) {
	// Prepare add liquidity parameters
	addLiquidityParams := map[string]interface{}{
		"currency0": ethereum.Token0_address,
		"currency1": ethereum.Token1_address,
		"amount0":   "1000000000000000000", // 1 of token0
		"amount1":   "1000000000000000000", // 1 of token1
		"minTick":   "-887220",             // Example value, adjust as needed
		"maxTick":   "887220",              // Example value, adjust as needed
	}

	jsonParams, err := json.Marshal(addLiquidityParams)
	assert.NoError(t, err)

	// Send POST request to /addLiquidity
	resp, err := http.Post(testServer.URL+"/addLiquidity", "application/json", bytes.NewBuffer(jsonParams))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Parse the response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	// Assert the response contains expected fields
	assert.Contains(t, result, "txHash")
	assert.Contains(t, result, "status")
	assert.Contains(t, result, "balancesBefore")
	assert.Contains(t, result, "balancesAfter")
	assert.Contains(t, result, "deltaBalances")
}
