// cmd/cli/main.go

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	baseURL = "http://localhost:8080" // Adjust this to match your server's address
)

func main() {
	approveCmd := flag.NewFlagSet("approve", flag.ExitOnError)

	initializeCmd := flag.NewFlagSet("initialize", flag.ExitOnError)
	initCurrency0 := initializeCmd.String("currency0", "", "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	initCurrency1 := initializeCmd.String("currency1", "", "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")

	approveCurrency0 := approveCmd.String("currency0", "", "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	approveCurrency1 := approveCmd.String("currency1", "", "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")

	addLiquidityCmd := flag.NewFlagSet("addLiquidity", flag.ExitOnError)
	addLiquidityCurrency0 := addLiquidityCmd.String("currency0", "", "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	addLiquidityCurrency1 := addLiquidityCmd.String("currency1", "", "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")

	addLiquidityPermitCmd := flag.NewFlagSet("addLiquidityPermit", flag.ExitOnError)

	swapCmd := flag.NewFlagSet("swap", flag.ExitOnError)
	swapCurrency0 := swapCmd.String("currency0", "", "0x0165878A594ca255338adfa4d48449f69242Eb8F")
	swapCurrency1 := swapCmd.String("currency1", "", "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707")

	swapAmount := swapCmd.String("amount", "", "100000000000000000")
	swapZeroForOne := swapCmd.Bool("zeroForOne", true, "Direction of swap (true for currency0 to currency1)")

	swapPermitCmd := flag.NewFlagSet("swapPermit", flag.ExitOnError)
	swapPermitCurrency0 := swapPermitCmd.String("currency0", "", "Address of currency0")
	swapPermitCurrency1 := swapPermitCmd.String("currency1", "", "Address of currency1")
	swapPermitAmount := swapPermitCmd.String("amount", "", "Amount to swap")
	swapPermitZeroForOne := swapPermitCmd.Bool("zeroForOne", true, "Direction of swap (true for currency0 to currency1)")
	swapPermitUserAddress := swapPermitCmd.String("userAddress", "", "User's address")
	swapPermitPrivateKey := swapPermitCmd.String("privateKey", "", "User's private key")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'approve', 'initialize', 'addLiquidity', 'addLiquidityPermit', 'swap', or 'swapPermit' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "approve":
		approveCmd.Parse(os.Args[2:])
		approve(*approveCurrency0, *approveCurrency1)
	case "initialize":
		initializeCmd.Parse(os.Args[2:])
		initialize(*initCurrency0, *initCurrency1)
	case "addLiquidity":
		addLiquidityCmd.Parse(os.Args[2:])
		addLiquidity(*addLiquidityCurrency0, *addLiquidityCurrency1)
	case "addLiquidityPermit":
		addLiquidityPermitCmd.Parse(os.Args[2:])
		addLiquidityPermit()
	case "swap":
		swapCmd.Parse(os.Args[2:])
		swap(*swapCurrency0, *swapCurrency1, *swapAmount, *swapZeroForOne)
	case "swapPermit":
		swapPermitCmd.Parse(os.Args[2:])
		swapPermit(*swapPermitCurrency0, *swapPermitCurrency1, *swapPermitAmount, *swapPermitZeroForOne, *swapPermitUserAddress, *swapPermitPrivateKey)
	default:
		fmt.Println("Expected 'approve', 'initialize', 'addLiquidity', 'addLiquidityPermit', 'swap', or 'swapPermit' subcommands")
		os.Exit(1)
	}
}

// Update the approve function:
func approve(currency0, currency1 string) {
	if currency0 == "" || currency1 == "" {
		fmt.Println("Both currency0 and currency1 addresses are required")
		return
	}

	requestBody, _ := json.Marshal(map[string]string{
		"currency0": currency0,
		"currency1": currency1,
	})

	resp, err := makeRequest("/approve", requestBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	fmt.Printf("Approve result:\n")
	for currency, data := range result {
		fmt.Printf("  %s:\n", currency)
		if dataMap, ok := data.(map[string]interface{}); ok {
			fmt.Printf("    Message: %s\n", dataMap["message"])
			fmt.Printf("    Balance: %s\n", dataMap["balance"])
		} else {
			fmt.Printf("    %v\n", data)
		}
	}
}

func swapPermit(currency0, currency1, amount string, zeroForOne bool, userAddress, privateKey string) {
	if currency0 == "" || currency1 == "" || amount == "" || userAddress == "" || privateKey == "" {
		fmt.Println("All parameters are required for swapPermit")
		return
	}

	requestBody, _ := json.Marshal(map[string]interface{}{
		"currency0":   currency0,
		"currency1":   currency1,
		"amount":      amount,
		"zeroForOne":  zeroForOne,
		"userAddress": userAddress,
		"privateKey":  privateKey,
	})

	resp, err := makeRequest("/performSwapWithPermit", requestBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print raw response
	fmt.Printf("Raw response: %s\n", string(resp))

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	fmt.Printf("Swap Permit result:\n")
	fmt.Printf("  Transaction Hash: %v\n", result["txHash"])
	fmt.Printf("  Message: %v\n", result["message"])
	fmt.Printf("  Balances Before: %v\n", result["balancesBefore"])
	fmt.Printf("  Balances After: %v\n", result["balancesAfter"])
	fmt.Printf("  Delta Balances: %v\n", result["deltaBalances"])
}

func initialize(currency0, currency1 string) {
	if currency0 == "" || currency1 == "" {
		fmt.Println("Both currency0 and currency1 addresses are required")
		return
	}

	requestBody, _ := json.Marshal(map[string]string{
		"currency0": currency0,
		"currency1": currency1,
	})

	resp, err := makeRequest("/initialize", requestBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	var result map[string]interface{}
	json.Unmarshal(resp, &result)

	fmt.Printf("Initialization result:\n")
	fmt.Printf("  Transaction Hash: %s\n", result["initializeTxHash"])
	fmt.Printf("  Status: %s\n", result["status"])
}

func addLiquidity(currency0, currency1 string) {
	if currency0 == "" || currency1 == "" {
		fmt.Println("Both currency0 and currency1 addresses are required")
		return
	}

	requestBody, _ := json.Marshal(map[string]string{
		"currency0": currency0,
		"currency1": currency1,
	})

	resp, err := makeRequest("/addLiquidity", requestBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print raw response
	fmt.Printf("Raw response: %s\n", string(resp))

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	fmt.Printf("Add Liquidity result:\n")
	fmt.Printf("  Status: %v\n", result["status"])
	fmt.Printf("  Transaction Hash: %v\n", result["txHash"])
	fmt.Printf("  Balances Before: %v\n", result["balancesBefore"])
	fmt.Printf("  Balances After: %v\n", result["balancesAfter"])
	fmt.Printf("  Delta Balances: %v\n", result["deltaBalances"])
}

func addLiquidityPermit() {
	fmt.Println("AddLiquidityPermit command not yet implemented")
}

func swap(currency0, currency1, amount string, zeroForOne bool) {
	if currency0 == "" || currency1 == "" || amount == "" {
		fmt.Println("Currency0, currency1, and amount are required")
		return
	}

	requestBody, _ := json.Marshal(map[string]interface{}{
		"currency0":  currency0,
		"currency1":  currency1,
		"amount":     amount,
		"zeroForOne": zeroForOne,
	})

	resp, err := makeRequest("/performSwap", requestBody)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print raw response
	fmt.Printf("Raw response: %s\n", string(resp))

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		return
	}

	fmt.Printf("Swap result:\n")
	fmt.Printf("  Transaction Hash: %v\n", result["txHash"])
	fmt.Printf("  Balances Before: %v\n", result["balancesBefore"])
	fmt.Printf("  Balances After: %v\n", result["balancesAfter"])
	fmt.Printf("  Delta Balances: %v\n", result["deltaBalances"])
}

func makeRequest(endpoint string, body []byte) ([]byte, error) {
	resp, err := http.Post(baseURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
