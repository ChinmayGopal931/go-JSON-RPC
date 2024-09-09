package ethereum

import (
	"strings"
	"uniswap-v4-rpc/internal/config"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	SwapRouterAddress common.Address
	LPRouterAddress   common.Address
	ManagerAddress    common.Address
	HookAddress       common.Address
	Token0_address    common.Address
	Token1_address    common.Address
	SwapRouterABI     abi.ABI
	LPRouterABI       abi.ABI
	ManagerABI        abi.ABI
)

func InitContracts(cfg *config.Config) error {
	var err error

	SwapRouterABI, err = abi.JSON(strings.NewReader(SwapRouterABIJSON))
	if err != nil {
		return err
	}

	LPRouterABI, err = abi.JSON(strings.NewReader(LPRouterABIJSON))
	if err != nil {
		return err
	}

	ManagerABI, err = abi.JSON(strings.NewReader(ManagerABIJSON))
	if err != nil {
		return err
	}

	SwapRouterAddress = common.HexToAddress(cfg.SwapRouterAddress)
	LPRouterAddress = common.HexToAddress(cfg.LPRouterAddress)
	ManagerAddress = common.HexToAddress(cfg.ManagerAddress)
	HookAddress = common.HexToAddress(cfg.HookAddress)

	//For testing
	Token0_address = common.HexToAddress(cfg.Token0_address)
	Token1_address = common.HexToAddress(cfg.Token1_address)

	return nil
}

// @DEV TODO HANDLE THIS BETTER
const SwapRouterABIJSON = `[
  {
    "type": "constructor",
    "inputs": [
      {
        "name": "_manager",
        "type": "address",
        "internalType": "contract IPoolManager"
      }
    ],
    "stateMutability": "nonpayable"
  },
  {
    "type": "function",
    "name": "SWAP_TYPEHASH",
    "inputs": [],
    "outputs": [{ "name": "", "type": "bytes32", "internalType": "bytes32" }],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "eip712Domain",
    "inputs": [],
    "outputs": [
      { "name": "fields", "type": "bytes1", "internalType": "bytes1" },
      { "name": "name", "type": "string", "internalType": "string" },
      { "name": "version", "type": "string", "internalType": "string" },
      { "name": "chainId", "type": "uint256", "internalType": "uint256" },
      {
        "name": "verifyingContract",
        "type": "address",
        "internalType": "address"
      },
      { "name": "salt", "type": "bytes32", "internalType": "bytes32" },
      {
        "name": "extensions",
        "type": "uint256[]",
        "internalType": "uint256[]"
      }
    ],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "manager",
    "inputs": [],
    "outputs": [
      {
        "name": "",
        "type": "address",
        "internalType": "contract IPoolManager"
      }
    ],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "nonces",
    "inputs": [{ "name": "", "type": "address", "internalType": "address" }],
    "outputs": [{ "name": "", "type": "uint256", "internalType": "uint256" }],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "swap",
    "inputs": [
      {
        "name": "key",
        "type": "tuple",
        "internalType": "struct PoolKey",
        "components": [
          {
            "name": "currency0",
            "type": "address",
            "internalType": "Currency"
          },
          {
            "name": "currency1",
            "type": "address",
            "internalType": "Currency"
          },
          { "name": "fee", "type": "uint24", "internalType": "uint24" },
          { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
          {
            "name": "hooks",
            "type": "address",
            "internalType": "contract IHooks"
          }
        ]
      },
      {
        "name": "params",
        "type": "tuple",
        "internalType": "struct IPoolManager.SwapParams",
        "components": [
          { "name": "zeroForOne", "type": "bool", "internalType": "bool" },
          {
            "name": "amountSpecified",
            "type": "int256",
            "internalType": "int256"
          },
          {
            "name": "sqrtPriceLimitX96",
            "type": "uint160",
            "internalType": "uint160"
          }
        ]
      },
      {
        "name": "testSettings",
        "type": "tuple",
        "internalType": "struct PoolSwapTest.TestSettings",
        "components": [
          { "name": "takeClaims", "type": "bool", "internalType": "bool" },
          {
            "name": "settleUsingBurn",
            "type": "bool",
            "internalType": "bool"
          }
        ]
      },
      { "name": "hookData", "type": "bytes", "internalType": "bytes" }
    ],
    "outputs": [
      { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
    ],
    "stateMutability": "payable"
  },
  {
    "type": "function",
    "name": "swapWithPermit",
    "inputs": [
      { "name": "user", "type": "address", "internalType": "address" },
      {
        "name": "key",
        "type": "tuple",
        "internalType": "struct PoolKey",
        "components": [
          {
            "name": "currency0",
            "type": "address",
            "internalType": "Currency"
          },
          {
            "name": "currency1",
            "type": "address",
            "internalType": "Currency"
          },
          { "name": "fee", "type": "uint24", "internalType": "uint24" },
          { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
          {
            "name": "hooks",
            "type": "address",
            "internalType": "contract IHooks"
          }
        ]
      },
      {
        "name": "params",
        "type": "tuple",
        "internalType": "struct IPoolManager.SwapParams",
        "components": [
          { "name": "zeroForOne", "type": "bool", "internalType": "bool" },
          {
            "name": "amountSpecified",
            "type": "int256",
            "internalType": "int256"
          },
          {
            "name": "sqrtPriceLimitX96",
            "type": "uint160",
            "internalType": "uint160"
          }
        ]
      },
      {
        "name": "testSettings",
        "type": "tuple",
        "internalType": "struct PoolSwapTest.TestSettings",
        "components": [
          { "name": "takeClaims", "type": "bool", "internalType": "bool" },
          {
            "name": "settleUsingBurn",
            "type": "bool",
            "internalType": "bool"
          }
        ]
      },
      { "name": "hookData", "type": "bytes", "internalType": "bytes" },
      { "name": "deadline", "type": "uint256", "internalType": "uint256" },
      { "name": "v", "type": "uint8", "internalType": "uint8" },
      { "name": "r", "type": "bytes32", "internalType": "bytes32" },
      { "name": "s", "type": "bytes32", "internalType": "bytes32" }
    ],
    "outputs": [
      { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
    ],
    "stateMutability": "payable"
  },
  {
    "type": "function",
    "name": "unlockCallback",
    "inputs": [
      { "name": "rawData", "type": "bytes", "internalType": "bytes" }
    ],
    "outputs": [{ "name": "", "type": "bytes", "internalType": "bytes" }],
    "stateMutability": "nonpayable"
  },
  {
    "type": "event",
    "name": "EIP712DomainChanged",
    "inputs": [],
    "anonymous": false
  },
  { "type": "error", "name": "InvalidShortString", "inputs": [] },
  { "type": "error", "name": "NoSwapOccurred", "inputs": [] },
  {
    "type": "error",
    "name": "StringTooLong",
    "inputs": [{ "name": "str", "type": "string", "internalType": "string" }]
  }
]`

const LPRouterABIJSON = `[
  {
    "type": "constructor",
    "inputs": [
      {
        "name": "_manager",
        "type": "address",
        "internalType": "contract IPoolManager"
      }
    ],
    "stateMutability": "nonpayable"
  },
  {
    "type": "function",
    "name": "MODIFY_LIQUIDITY_TYPEHASH",
    "inputs": [],
    "outputs": [{ "name": "", "type": "bytes32", "internalType": "bytes32" }],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "eip712Domain",
    "inputs": [],
    "outputs": [
      { "name": "fields", "type": "bytes1", "internalType": "bytes1" },
      { "name": "name", "type": "string", "internalType": "string" },
      { "name": "version", "type": "string", "internalType": "string" },
      { "name": "chainId", "type": "uint256", "internalType": "uint256" },
      {
        "name": "verifyingContract",
        "type": "address",
        "internalType": "address"
      },
      { "name": "salt", "type": "bytes32", "internalType": "bytes32" },
      {
        "name": "extensions",
        "type": "uint256[]",
        "internalType": "uint256[]"
      }
    ],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "manager",
    "inputs": [],
    "outputs": [
      {
        "name": "",
        "type": "address",
        "internalType": "contract IPoolManager"
      }
    ],
    "stateMutability": "view"
  },
  {
    "type": "function",
    "name": "modifyLiquidity",
    "inputs": [
      {
        "name": "key",
        "type": "tuple",
        "internalType": "struct PoolKey",
        "components": [
          {
            "name": "currency0",
            "type": "address",
            "internalType": "Currency"
          },
          {
            "name": "currency1",
            "type": "address",
            "internalType": "Currency"
          },
          { "name": "fee", "type": "uint24", "internalType": "uint24" },
          { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
          {
            "name": "hooks",
            "type": "address",
            "internalType": "contract IHooks"
          }
        ]
      },
      {
        "name": "params",
        "type": "tuple",
        "internalType": "struct IPoolManager.ModifyLiquidityParams",
        "components": [
          { "name": "tickLower", "type": "int24", "internalType": "int24" },
          { "name": "tickUpper", "type": "int24", "internalType": "int24" },
          {
            "name": "liquidityDelta",
            "type": "int256",
            "internalType": "int256"
          },
          { "name": "salt", "type": "bytes32", "internalType": "bytes32" }
        ]
      },
      { "name": "hookData", "type": "bytes", "internalType": "bytes" },
      { "name": "settleUsingBurn", "type": "bool", "internalType": "bool" },
      { "name": "takeClaims", "type": "bool", "internalType": "bool" }
    ],
    "outputs": [
      { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
    ],
    "stateMutability": "payable"
  },
  {
    "type": "function",
    "name": "modifyLiquidity",
    "inputs": [
      {
        "name": "key",
        "type": "tuple",
        "internalType": "struct PoolKey",
        "components": [
          {
            "name": "currency0",
            "type": "address",
            "internalType": "Currency"
          },
          {
            "name": "currency1",
            "type": "address",
            "internalType": "Currency"
          },
          { "name": "fee", "type": "uint24", "internalType": "uint24" },
          { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
          {
            "name": "hooks",
            "type": "address",
            "internalType": "contract IHooks"
          }
        ]
      },
      {
        "name": "params",
        "type": "tuple",
        "internalType": "struct IPoolManager.ModifyLiquidityParams",
        "components": [
          { "name": "tickLower", "type": "int24", "internalType": "int24" },
          { "name": "tickUpper", "type": "int24", "internalType": "int24" },
          {
            "name": "liquidityDelta",
            "type": "int256",
            "internalType": "int256"
          },
          { "name": "salt", "type": "bytes32", "internalType": "bytes32" }
        ]
      },
      { "name": "hookData", "type": "bytes", "internalType": "bytes" }
    ],
    "outputs": [
      { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
    ],
    "stateMutability": "payable"
  },
  {
    "type": "function",
    "name": "modifyLiquidityWithPermit",
    "inputs": [
      { "name": "user", "type": "address", "internalType": "address" },
      {
        "name": "key",
        "type": "tuple",
        "internalType": "struct PoolKey",
        "components": [
          {
            "name": "currency0",
            "type": "address",
            "internalType": "Currency"
          },
          {
            "name": "currency1",
            "type": "address",
            "internalType": "Currency"
          },
          { "name": "fee", "type": "uint24", "internalType": "uint24" },
          { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
          {
            "name": "hooks",
            "type": "address",
            "internalType": "contract IHooks"
          }
        ]
      },
      {
        "name": "params",
        "type": "tuple",
        "internalType": "struct IPoolManager.ModifyLiquidityParams",
        "components": [
          { "name": "tickLower", "type": "int24", "internalType": "int24" },
          { "name": "tickUpper", "type": "int24", "internalType": "int24" },
          {
            "name": "liquidityDelta",
            "type": "int256",
            "internalType": "int256"
          },
          { "name": "salt", "type": "bytes32", "internalType": "bytes32" }
        ]
      },
      { "name": "hookData", "type": "bytes", "internalType": "bytes" },
      { "name": "settleUsingBurn", "type": "bool", "internalType": "bool" },
      { "name": "takeClaims", "type": "bool", "internalType": "bool" },
      { "name": "deadline", "type": "uint256", "internalType": "uint256" },
      { "name": "v0", "type": "uint8", "internalType": "uint8" },
      { "name": "r0", "type": "bytes32", "internalType": "bytes32" },
      { "name": "s0", "type": "bytes32", "internalType": "bytes32" },
      { "name": "v1", "type": "uint8", "internalType": "uint8" },
      { "name": "r1", "type": "bytes32", "internalType": "bytes32" },
      { "name": "s1", "type": "bytes32", "internalType": "bytes32" }
    ],
    "outputs": [
      { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
    ],
    "stateMutability": "payable"
  },
  {
    "type": "function",
    "name": "unlockCallback",
    "inputs": [
      { "name": "rawData", "type": "bytes", "internalType": "bytes" }
    ],
    "outputs": [{ "name": "", "type": "bytes", "internalType": "bytes" }],
    "stateMutability": "nonpayable"
  },
  {
    "type": "event",
    "name": "EIP712DomainChanged",
    "inputs": [],
    "anonymous": false
  },
  { "type": "error", "name": "InvalidShortString", "inputs": [] },
  {
    "type": "error",
    "name": "StringTooLong",
    "inputs": [{ "name": "str", "type": "string", "internalType": "string" }]
  }
]`

const ManagerABIJSON = `[
    {
      "type": "constructor",
      "inputs": [
        {
          "name": "controllerGasLimit",
          "type": "uint256",
          "internalType": "uint256"
        }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "allowance",
      "inputs": [
        { "name": "owner", "type": "address", "internalType": "address" },
        { "name": "spender", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "approve",
      "inputs": [
        { "name": "spender", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [{ "name": "", "type": "bool", "internalType": "bool" }],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "balanceOf",
      "inputs": [
        { "name": "owner", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [
        { "name": "balance", "type": "uint256", "internalType": "uint256" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "burn",
      "inputs": [
        { "name": "from", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "clear",
      "inputs": [
        { "name": "currency", "type": "address", "internalType": "Currency" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "collectProtocolFees",
      "inputs": [
        { "name": "recipient", "type": "address", "internalType": "address" },
        { "name": "currency", "type": "address", "internalType": "Currency" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [
        {
          "name": "amountCollected",
          "type": "uint256",
          "internalType": "uint256"
        }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "donate",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        { "name": "amount0", "type": "uint256", "internalType": "uint256" },
        { "name": "amount1", "type": "uint256", "internalType": "uint256" },
        { "name": "hookData", "type": "bytes", "internalType": "bytes" }
      ],
      "outputs": [
        { "name": "delta", "type": "int256", "internalType": "BalanceDelta" }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "extsload",
      "inputs": [
        { "name": "slot", "type": "bytes32", "internalType": "bytes32" }
      ],
      "outputs": [{ "name": "", "type": "bytes32", "internalType": "bytes32" }],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "extsload",
      "inputs": [
        { "name": "startSlot", "type": "bytes32", "internalType": "bytes32" },
        { "name": "nSlots", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [
        { "name": "", "type": "bytes32[]", "internalType": "bytes32[]" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "extsload",
      "inputs": [
        { "name": "slots", "type": "bytes32[]", "internalType": "bytes32[]" }
      ],
      "outputs": [
        { "name": "", "type": "bytes32[]", "internalType": "bytes32[]" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "exttload",
      "inputs": [
        { "name": "slots", "type": "bytes32[]", "internalType": "bytes32[]" }
      ],
      "outputs": [
        { "name": "", "type": "bytes32[]", "internalType": "bytes32[]" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "exttload",
      "inputs": [
        { "name": "slot", "type": "bytes32", "internalType": "bytes32" }
      ],
      "outputs": [{ "name": "", "type": "bytes32", "internalType": "bytes32" }],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "initialize",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        {
          "name": "sqrtPriceX96",
          "type": "uint160",
          "internalType": "uint160"
        },
        { "name": "hookData", "type": "bytes", "internalType": "bytes" }
      ],
      "outputs": [{ "name": "tick", "type": "int24", "internalType": "int24" }],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "isOperator",
      "inputs": [
        { "name": "owner", "type": "address", "internalType": "address" },
        { "name": "operator", "type": "address", "internalType": "address" }
      ],
      "outputs": [
        { "name": "isOperator", "type": "bool", "internalType": "bool" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "mint",
      "inputs": [
        { "name": "to", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "modifyLiquidity",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        {
          "name": "params",
          "type": "tuple",
          "internalType": "struct IPoolManager.ModifyLiquidityParams",
          "components": [
            { "name": "tickLower", "type": "int24", "internalType": "int24" },
            { "name": "tickUpper", "type": "int24", "internalType": "int24" },
            {
              "name": "liquidityDelta",
              "type": "int256",
              "internalType": "int256"
            },
            { "name": "salt", "type": "bytes32", "internalType": "bytes32" }
          ]
        },
        { "name": "hookData", "type": "bytes", "internalType": "bytes" }
      ],
      "outputs": [
        {
          "name": "callerDelta",
          "type": "int256",
          "internalType": "BalanceDelta"
        },
        {
          "name": "feesAccrued",
          "type": "int256",
          "internalType": "BalanceDelta"
        }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "owner",
      "inputs": [],
      "outputs": [{ "name": "", "type": "address", "internalType": "address" }],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "protocolFeeController",
      "inputs": [],
      "outputs": [
        {
          "name": "",
          "type": "address",
          "internalType": "contract IProtocolFeeController"
        }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "protocolFeesAccrued",
      "inputs": [
        { "name": "currency", "type": "address", "internalType": "Currency" }
      ],
      "outputs": [
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "setOperator",
      "inputs": [
        { "name": "operator", "type": "address", "internalType": "address" },
        { "name": "approved", "type": "bool", "internalType": "bool" }
      ],
      "outputs": [{ "name": "", "type": "bool", "internalType": "bool" }],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "setProtocolFee",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        { "name": "newProtocolFee", "type": "uint24", "internalType": "uint24" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "setProtocolFeeController",
      "inputs": [
        {
          "name": "controller",
          "type": "address",
          "internalType": "contract IProtocolFeeController"
        }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "settle",
      "inputs": [],
      "outputs": [
        { "name": "paid", "type": "uint256", "internalType": "uint256" }
      ],
      "stateMutability": "payable"
    },
    {
      "type": "function",
      "name": "settleFor",
      "inputs": [
        { "name": "recipient", "type": "address", "internalType": "address" }
      ],
      "outputs": [
        { "name": "paid", "type": "uint256", "internalType": "uint256" }
      ],
      "stateMutability": "payable"
    },
    {
      "type": "function",
      "name": "supportsInterface",
      "inputs": [
        { "name": "interfaceId", "type": "bytes4", "internalType": "bytes4" }
      ],
      "outputs": [{ "name": "", "type": "bool", "internalType": "bool" }],
      "stateMutability": "view"
    },
    {
      "type": "function",
      "name": "swap",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        {
          "name": "params",
          "type": "tuple",
          "internalType": "struct IPoolManager.SwapParams",
          "components": [
            { "name": "zeroForOne", "type": "bool", "internalType": "bool" },
            {
              "name": "amountSpecified",
              "type": "int256",
              "internalType": "int256"
            },
            {
              "name": "sqrtPriceLimitX96",
              "type": "uint160",
              "internalType": "uint160"
            }
          ]
        },
        { "name": "hookData", "type": "bytes", "internalType": "bytes" }
      ],
      "outputs": [
        {
          "name": "swapDelta",
          "type": "int256",
          "internalType": "BalanceDelta"
        }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "sync",
      "inputs": [
        { "name": "currency", "type": "address", "internalType": "Currency" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "take",
      "inputs": [
        { "name": "currency", "type": "address", "internalType": "Currency" },
        { "name": "to", "type": "address", "internalType": "address" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "transfer",
      "inputs": [
        { "name": "receiver", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [{ "name": "", "type": "bool", "internalType": "bool" }],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "transferFrom",
      "inputs": [
        { "name": "sender", "type": "address", "internalType": "address" },
        { "name": "receiver", "type": "address", "internalType": "address" },
        { "name": "id", "type": "uint256", "internalType": "uint256" },
        { "name": "amount", "type": "uint256", "internalType": "uint256" }
      ],
      "outputs": [{ "name": "", "type": "bool", "internalType": "bool" }],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "transferOwnership",
      "inputs": [
        { "name": "newOwner", "type": "address", "internalType": "address" }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "unlock",
      "inputs": [{ "name": "data", "type": "bytes", "internalType": "bytes" }],
      "outputs": [
        { "name": "result", "type": "bytes", "internalType": "bytes" }
      ],
      "stateMutability": "nonpayable"
    },
    {
      "type": "function",
      "name": "updateDynamicLPFee",
      "inputs": [
        {
          "name": "key",
          "type": "tuple",
          "internalType": "struct PoolKey",
          "components": [
            {
              "name": "currency0",
              "type": "address",
              "internalType": "Currency"
            },
            {
              "name": "currency1",
              "type": "address",
              "internalType": "Currency"
            },
            { "name": "fee", "type": "uint24", "internalType": "uint24" },
            { "name": "tickSpacing", "type": "int24", "internalType": "int24" },
            {
              "name": "hooks",
              "type": "address",
              "internalType": "contract IHooks"
            }
          ]
        },
        {
          "name": "newDynamicLPFee",
          "type": "uint24",
          "internalType": "uint24"
        }
      ],
      "outputs": [],
      "stateMutability": "nonpayable"
    },
    {
      "type": "event",
      "name": "Approval",
      "inputs": [
        {
          "name": "owner",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "spender",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "id",
          "type": "uint256",
          "indexed": true,
          "internalType": "uint256"
        },
        {
          "name": "amount",
          "type": "uint256",
          "indexed": false,
          "internalType": "uint256"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "Initialize",
      "inputs": [
        {
          "name": "id",
          "type": "bytes32",
          "indexed": true,
          "internalType": "PoolId"
        },
        {
          "name": "currency0",
          "type": "address",
          "indexed": true,
          "internalType": "Currency"
        },
        {
          "name": "currency1",
          "type": "address",
          "indexed": true,
          "internalType": "Currency"
        },
        {
          "name": "fee",
          "type": "uint24",
          "indexed": false,
          "internalType": "uint24"
        },
        {
          "name": "tickSpacing",
          "type": "int24",
          "indexed": false,
          "internalType": "int24"
        },
        {
          "name": "hooks",
          "type": "address",
          "indexed": false,
          "internalType": "contract IHooks"
        },
        {
          "name": "sqrtPriceX96",
          "type": "uint160",
          "indexed": false,
          "internalType": "uint160"
        },
        {
          "name": "tick",
          "type": "int24",
          "indexed": false,
          "internalType": "int24"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "ModifyLiquidity",
      "inputs": [
        {
          "name": "id",
          "type": "bytes32",
          "indexed": true,
          "internalType": "PoolId"
        },
        {
          "name": "sender",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "tickLower",
          "type": "int24",
          "indexed": false,
          "internalType": "int24"
        },
        {
          "name": "tickUpper",
          "type": "int24",
          "indexed": false,
          "internalType": "int24"
        },
        {
          "name": "liquidityDelta",
          "type": "int256",
          "indexed": false,
          "internalType": "int256"
        },
        {
          "name": "salt",
          "type": "bytes32",
          "indexed": false,
          "internalType": "bytes32"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "OperatorSet",
      "inputs": [
        {
          "name": "owner",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "operator",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "approved",
          "type": "bool",
          "indexed": false,
          "internalType": "bool"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "OwnershipTransferred",
      "inputs": [
        {
          "name": "user",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "newOwner",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "ProtocolFeeControllerUpdated",
      "inputs": [
        {
          "name": "protocolFeeController",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "ProtocolFeeUpdated",
      "inputs": [
        {
          "name": "id",
          "type": "bytes32",
          "indexed": true,
          "internalType": "PoolId"
        },
        {
          "name": "protocolFee",
          "type": "uint24",
          "indexed": false,
          "internalType": "uint24"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "Swap",
      "inputs": [
        {
          "name": "id",
          "type": "bytes32",
          "indexed": true,
          "internalType": "PoolId"
        },
        {
          "name": "sender",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "amount0",
          "type": "int128",
          "indexed": false,
          "internalType": "int128"
        },
        {
          "name": "amount1",
          "type": "int128",
          "indexed": false,
          "internalType": "int128"
        },
        {
          "name": "sqrtPriceX96",
          "type": "uint160",
          "indexed": false,
          "internalType": "uint160"
        },
        {
          "name": "liquidity",
          "type": "uint128",
          "indexed": false,
          "internalType": "uint128"
        },
        {
          "name": "tick",
          "type": "int24",
          "indexed": false,
          "internalType": "int24"
        },
        {
          "name": "fee",
          "type": "uint24",
          "indexed": false,
          "internalType": "uint24"
        }
      ],
      "anonymous": false
    },
    {
      "type": "event",
      "name": "Transfer",
      "inputs": [
        {
          "name": "caller",
          "type": "address",
          "indexed": false,
          "internalType": "address"
        },
        {
          "name": "from",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "to",
          "type": "address",
          "indexed": true,
          "internalType": "address"
        },
        {
          "name": "id",
          "type": "uint256",
          "indexed": true,
          "internalType": "uint256"
        },
        {
          "name": "amount",
          "type": "uint256",
          "indexed": false,
          "internalType": "uint256"
        }
      ],
      "anonymous": false
    },
    { "type": "error", "name": "AlreadyUnlocked", "inputs": [] },
    {
      "type": "error",
      "name": "CurrenciesOutOfOrderOrEqual",
      "inputs": [
        { "name": "currency0", "type": "address", "internalType": "address" },
        { "name": "currency1", "type": "address", "internalType": "address" }
      ]
    },
    { "type": "error", "name": "CurrencyNotSettled", "inputs": [] },
    { "type": "error", "name": "DelegateCallNotAllowed", "inputs": [] },
    { "type": "error", "name": "InvalidCaller", "inputs": [] },
    { "type": "error", "name": "ManagerLocked", "inputs": [] },
    { "type": "error", "name": "MustClearExactPositiveDelta", "inputs": [] },
    { "type": "error", "name": "NonzeroNativeValue", "inputs": [] },
    { "type": "error", "name": "PoolNotInitialized", "inputs": [] },
    { "type": "error", "name": "ProtocolFeeCannotBeFetched", "inputs": [] },
    {
      "type": "error",
      "name": "ProtocolFeeTooLarge",
      "inputs": [{ "name": "fee", "type": "uint24", "internalType": "uint24" }]
    },
    { "type": "error", "name": "SwapAmountCannotBeZero", "inputs": [] },
    {
      "type": "error",
      "name": "TickSpacingTooLarge",
      "inputs": [
        { "name": "tickSpacing", "type": "int24", "internalType": "int24" }
      ]
    },
    {
      "type": "error",
      "name": "TickSpacingTooSmall",
      "inputs": [
        { "name": "tickSpacing", "type": "int24", "internalType": "int24" }
      ]
    },
    { "type": "error", "name": "UnauthorizedDynamicLPFeeUpdate", "inputs": [] }
  ]`
