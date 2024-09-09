# Uniswap V4 JSON-RPC Server

  

This project implements a JSON-RPC 2.0 server for interacting with Uniswap V4 pools. It provides endpoints for swapping tokens, adding and removing liquidity, and other pool-related operations.

  

## Table of Contents

  

1. [Project Structure](#project-structure)

2. [Setup](#setup)

3. [Configuration](#configuration)

4. [Running the Server](#running-the-server)

5. [API Endpoints](#api-endpoints)

6. [CLI Tool](#cli-tool)

7. [Testing](#testing)

8. [Examples](#examples)

  

## Project Structure

`

uniswap-v4-rpc/

│

├── cmd/

│ ├── server/


│ └── main.go

│

├── internal/

│ ├── config/

│ │ └── config.go

│ ├── handlers/

│ │ ├── swap.go

│ │ ├── liquidity.go

│ │ └── initialize.go

│ ├── ethereum/

│ │ ├── client.go

│ │ └── contracts.go

│ └── routes/

│ └── routes.go

│

├── pkg/

│ └── utils/

│ └── ethereum_utils.go

│

├── test/

│ └── integration/

│ ├── setup_test.go

│ ├── swap_test.go

│ └── liquidity_test.go

│

├── config.yaml

├── go.mod

├── go.sum

└── README.md

`

  

## Setup

  

1. Clone the repository:
2. Run `Anvil` to spin up a local blockchain
3. Compile Contracts and Navigate to the contracts directory and run (this runs the deploy script)
`forge script script/Anvil.s.sol \
    --rpc-url http://localhost:8545 \
    --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
    --broadcast --via-ir  `
4. Read the logs from the deploy script and copy them over to the config.yaml file
5. Run `go run main.go` in the root Directory to start the server


  

## Configuration

  

The `config.yaml` file contains all necessary configuration for the server:



`# Ethereum Network Configuration

ethereum_node_url: "http://localhost:8545"

chain_id: 31337  # Local development chain ID

  

##### Contract Addresses

swap_router_address: "0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9"

lp_router_address: "0x9fE46736679d2D9a65F0992F2272dE9f3c7fa6e0"

manager_address: "0x5FbDB2315678afecb367f032d93F642f64180aa3"

hook_address: "0xA4B10483554041f45fe0E481B6Adc26b17eA0aC0"

  

###### Account Configuration

private_key: "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

  

###### API Server Configuration

server_host: "localhost"

server_port: 8080

  

###### Gas Configuration

gas_limit: 500000

gas_price: 20  # 20 Gwei

  

###### Token Addresses

token0_address: "0x0165878A594ca255338adfa4d48449f69242Eb8F"

token1_address: "0x5FC8d32690cc91D4c39d9d3abcBD16989F875707"

  

###### Pool Configuration

default_fee: 3000

default_tick_spacing: 60

  

###### Logging

log_level: "debug"

  

###### Environment

environment: "development"`

  
  

go run cmd/server/main.go

  
  

##API Endpoints

  

/approve: Approve tokens on the lp router and swap router address

/initialize: Initialize a new Uniswap V4 pool

/addLiquidity: Add liquidity to a pool

/removeLiquidity: Remove liquidity from a pool

/performSwap: Execute a token swap

/performSwapWithPermit: Execute a token swap with permit (ERC-2612)

/addLiquidityPermit: Execute modify liquidity with permit (ERC-2612)

  

## CLI Tool

The project includes a CLI tool for interacting with the JSON-RPC server. To build the CLI tool:
``go build -o uniswap-cli cmd/main.go``

Update the cmd/main.go file with the correct addresses. 

Example Usage
`./uniswap-cli approve -currency0 0x1234... -currency1 0x5678...

./uniswap-cli swap -currency0 0x1234... -currency1 0x5678... -amount 1000000000000000000 -zeroForOne=true
`
`./uniswap-cli addLiquidity -currency0 0x1234... -currency1 0x5678... -amount 1000000000000000000 -zeroForOne=true
`

## Testing

The project includes integration tests that interact with a local Ethereum testnet (e.g., Ganache, Hardhat).

To run the tests:

1.  Ensure your local Ethereum testnet is running.
2.  Update `config.yaml` in the test/integration folder with the contract details
3.  Run the tests:
    
  `go test -v ./test/integration/...`
    

The tests cover:

-   Address check (`address_check_test.go`)
-   Swapping tokens (`swap_test.go`)
-   Adding liquidity (`liquidity_test.go`) // Can remove liquidity if you update the value 
-   Initializing pools and other operations (`setup_test.go`)
