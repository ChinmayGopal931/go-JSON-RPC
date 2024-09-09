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

```
foundryup
```

## Set up

*requires [foundry](https://book.getfoundry.sh)*

```
forge install
forge test
```


1. Clone the repository:
2. Run `Anvil` to spin up a local blockchain
3. Compile Contracts and Navigate to the contracts directory and run (this runs the deploy script)

```bash
# start anvil, a local EVM chain
anvil

# in a new terminal
forge script script/Anvil.s.sol \
    --rpc-url http://localhost:8545 \
    --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
    --broadcast
```

4. Read the logs from the deploy script and copy them over to the config.yaml file
5. Run `go run main.go` in the root Directory to start the server

  ![Screenshot 2024-09-08 at 7 46 44 PM](https://github.com/user-attachments/assets/1090ee01-40c8-4d1d-9b96-beee086568d6)



## Configuration

  

The `config.yaml` file contains all necessary configuration for the server:


```
# Ethereum Network Configuration

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

  ```
  

go run cmd/server/main.go

  
  

##API Endpoints

  
```
/approve: Approve tokens on the lp router and swap router address

Example Usage

`
curl -X POST http://localhost:8080/approve \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourCurrency0Address",
  "currency1": "0xYourCurrency1Address"
}'
`

/initialize: Initialize a new Uniswap V4 pool

`
curl -X POST http://localhost:8080/initialize \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourCurrency0Address",
  "currency1": "0xYourCurrency1Address"
}'
`

/addLiquidity: Add liquidity to a pool

`
curl -X POST http://localhost:8080/addLiquidity \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourCurrency0Address",
  "currency1": "0xYourCurrency1Address"
}'
`


/performSwap: Execute a token swap


`
curl -X POST http://localhost:8080/perfromSwap \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourCurrency0Address",
  "currency1": "0xYourCurrency1Address",
  "amount": "1000000000000000000",  
  "zeroForOne": true
}'
`

/performSwapWithPermit: Execute a token swap with permit (ERC-2612)

`
curl -X POST http://localhost:8080/performSwapWithPermit \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourTokenAddress0",
  "currency1": "0xYourTokenAddress1",
  "amount": "1000000000000000000", 
  "zeroForOne": true,
  "userAddress": "0xYourEthereumAddress",
  "privateKey": "0xYourPrivateKey"
}'

`

/addLiquidityPermit: Execute modify liquidity with permit (ERC-2612)

`
curl -X POST http://localhost:8080/addLiquidityPermit \
-H "Content-Type: application/json" \
-d '{
  "currency0": "0xYourCurrency0Address",
  "currency1": "0xYourCurrency1Address",
  "amount": "1000000000000000000",  
  "zeroForOne": true
}'
`

```
  


## CLI Tool

The project includes a CLI tool for interacting with the JSON-RPC server. To build the CLI tool:
``go build -o uniswap-cli cmd/main.go``

Update the cmd/main.go file with the correct addresses. 

Example Usage

```./uniswap-cli approve -currency0 0x1234... -currency1 0x5678...```

```./uniswap-cli swap -currency0 0x1234... -currency1 0x5678... -amount 1000000000000000000 -zeroForOne=true```

```./uniswap-cli addLiquidity -currency0 0x1234... -currency1 0x5678... -amount 1000000000000000000 -zeroForOne=true```


## Testing

The project includes integration tests that interact with a local Ethereum blockchain Anvil.

To run the tests:

1.  Ensure your local Ethereum blockkchain testnet is running (`Anvil`).
2.  Update `config.yaml` in the test/integration folder with the contract details
3.  Run the golang tests:
    
  `go test -v ./test/integration/...`
4. Run Foundry Tests 
   `forge test`
    

The tests cover:

Server:
-   Address check (`address_check_test.go`)
-   Swapping tokens (`swap_test.go`)
-   Adding liquidity (`liquidity_test.go`) // Can remove liquidity if you update the value 
-   Initializing pools and other operations (`setup_test.go`)

Contracts:
-  (`Counter.t.sol`) Checks for correct ERC-2612 simplementation as well as simple hook functionality. 



