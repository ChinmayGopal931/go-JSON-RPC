package integration

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"
	"uniswap-v4-rpc/internal/ethereum"

	"github.com/stretchr/testify/assert"
)

func TestSwap(t *testing.T) {

	log.Println(ethereum.Token0_address, ethereum.Token1_address)
	swapParams := map[string]interface{}{
		"currency0":  ethereum.Token0_address,
		"currency1":  ethereum.Token1_address,
		"amount":     "1000000000",
		"zeroForOne": true,
	}

	jsonParams, err := json.Marshal(swapParams)
	assert.NoError(t, err)

	// Send POST request to /performSwap
	resp, err := http.Post(testServer.URL+"/performSwap", "application/json", bytes.NewBuffer(jsonParams))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	assert.NoError(t, err)

	//@dev this can be done better but works for now
	assert.Contains(t, result, "txHash")
	assert.Contains(t, result, "balancesBefore")
	assert.Contains(t, result, "balancesAfter")
	assert.Contains(t, result, "deltaBalances")

}
