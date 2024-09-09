// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "forge-std/Script.sol";

import {CurrencyLibrary, Currency} from "v4-core/src/types/Currency.sol";
import {IPoolManager} from "v4-core/src/interfaces/IPoolManager.sol";
import {BalanceDelta, BalanceDeltaLibrary} from "v4-core/src/types/BalanceDelta.sol";
import {PoolKey} from "v4-core/src/types/PoolKey.sol";
import {PoolId, PoolIdLibrary} from "@uniswap/v4-core/src/types/PoolId.sol";
import {PoolTestBase} from "@uniswap/v4-core/src/test/PoolTestBase.sol";
import {IHooks} from "v4-core/src/interfaces/IHooks.sol";
import {Hooks} from "v4-core/src/libraries/Hooks.sol";
import {LPFeeLibrary} from "v4-core/src/libraries/LPFeeLibrary.sol";
import {CurrencySettler} from "v4-core/test/utils/CurrencySettler.sol";
import {StateLibrary} from "v4-core/src/libraries/StateLibrary.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Permit.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";

contract PoolModifyLiquidityTest is PoolTestBase, EIP712 {
    using CurrencySettler for Currency;
    using Hooks for IHooks;
    using LPFeeLibrary for uint24;
    using PoolIdLibrary for PoolKey;
    using StateLibrary for IPoolManager;
    using ECDSA for bytes32;

    bytes32 public constant MODIFY_LIQUIDITY_TYPEHASH =
        keccak256("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)");

    constructor(IPoolManager _manager) PoolTestBase(_manager) EIP712("PoolModifyLiquidityTest", "1") {}

    struct CallbackData {
        address sender;
        PoolKey key;
        IPoolManager.ModifyLiquidityParams params;
        bytes hookData;
        bool settleUsingBurn;
        bool takeClaims;
    }

    function modifyLiquidity(
        PoolKey memory key,
        IPoolManager.ModifyLiquidityParams memory params,
        bytes memory hookData
    ) external payable returns (BalanceDelta delta) {
        delta = modifyLiquidity(key, params, hookData, false, false);
    }

    function modifyLiquidity(
        PoolKey memory key,
        IPoolManager.ModifyLiquidityParams memory params,
        bytes memory hookData,
        bool settleUsingBurn,
        bool takeClaims
    ) public payable returns (BalanceDelta delta) {
        delta = abi.decode(
            manager.unlock(abi.encode(CallbackData(msg.sender, key, params, hookData, settleUsingBurn, takeClaims))),
            (BalanceDelta)
        );

        uint256 ethBalance = address(this).balance;
        if (ethBalance > 0) {
            CurrencyLibrary.NATIVE.transfer(msg.sender, ethBalance);
        }
    }

    function modifyLiquidityWithPermit(
        address user,
        PoolKey memory key,
        IPoolManager.ModifyLiquidityParams memory params,
        bytes memory hookData,
        bool settleUsingBurn,
        bool takeClaims,
        uint256 deadline,
        uint8 v0,
        bytes32 r0,
        bytes32 s0,
        uint8 v1,
        bytes32 r1,
        bytes32 s1
    ) external payable returns (BalanceDelta delta) {
        IERC20Permit token0 = IERC20Permit(Currency.unwrap(key.currency0));
        IERC20Permit token1 = IERC20Permit(Currency.unwrap(key.currency1));

        uint256 amount0 = params.liquidityDelta > 0 ? uint256(params.liquidityDelta) : uint256(-params.liquidityDelta);
        uint256 amount1 = params.liquidityDelta > 0 ? uint256(params.liquidityDelta) : uint256(-params.liquidityDelta);

        uint256 currentNonce = token0.nonces(user);

        bytes32 domainSeparator = token0.DOMAIN_SEPARATOR();

        bytes32 permitHash =
            keccak256(abi.encode(MODIFY_LIQUIDITY_TYPEHASH, user, address(this), amount0, currentNonce, deadline));

        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, permitHash));

        address recoveredSigner = ecrecover(digest, v0, r0, s0);

        //@dev redundant test for testing purposes
        require(recoveredSigner != address(0) && recoveredSigner == user, "Invalid permit signature");

        token0.permit(user, address(this), amount0, deadline, v0, r0, s0);
        token1.permit(user, address(this), amount1, deadline, v1, r1, s1);

        // Proceed with modifying liquidity
        delta = abi.decode(
            manager.unlock(abi.encode(CallbackData(user, key, params, hookData, settleUsingBurn, takeClaims))),
            (BalanceDelta)
        );

        uint256 ethBalance = address(this).balance;
        if (ethBalance > 0) CurrencyLibrary.NATIVE.transfer(user, ethBalance);

        return delta;
    }

    function unlockCallback(bytes calldata rawData) external returns (bytes memory) {
        require(msg.sender == address(manager), "Caller must be manager");

        CallbackData memory data = abi.decode(rawData, (CallbackData));

        (uint128 liquidityBefore,,) = manager.getPositionInfo(
            data.key.toId(), address(this), data.params.tickLower, data.params.tickUpper, data.params.salt
        );

        (BalanceDelta delta,) = manager.modifyLiquidity(data.key, data.params, data.hookData);
        (uint128 liquidityAfter,,) = manager.getPositionInfo(
            data.key.toId(), address(this), data.params.tickLower, data.params.tickUpper, data.params.salt
        );
        (,, int256 delta0) = _fetchBalances(data.key.currency0, data.sender, address(this));
        (,, int256 delta1) = _fetchBalances(data.key.currency1, data.sender, address(this));

        require(
            int128(liquidityBefore) + data.params.liquidityDelta == int128(liquidityAfter), "Liquidity change incorrect"
        );

        if (data.params.liquidityDelta < 0) {
            assert(delta0 > 0 || delta1 > 0);
            assert(!(delta0 < 0 || delta1 < 0));
        } else if (data.params.liquidityDelta > 0) {
            assert(delta0 < 0 || delta1 < 0);
            assert(!(delta0 > 0 || delta1 > 0));
        }

        if (delta0 < 0) {
            data.key.currency0.settle(manager, data.sender, uint256(-delta0), data.settleUsingBurn);
        }
        if (delta1 < 0) {
            data.key.currency1.settle(manager, data.sender, uint256(-delta1), data.settleUsingBurn);
        }
        if (delta0 > 0) {
            data.key.currency0.take(manager, data.sender, uint256(delta0), data.takeClaims);
        }
        if (delta1 > 0) {
            data.key.currency1.take(manager, data.sender, uint256(delta1), data.takeClaims);
        }

        return abi.encode(delta);
    }
}
