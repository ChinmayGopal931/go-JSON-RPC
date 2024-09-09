// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

import "forge-std/Script.sol";

import {CurrencyLibrary, Currency} from "../types/Currency.sol";
import {IPoolManager} from "../interfaces/IPoolManager.sol";
import {BalanceDelta} from "../types/BalanceDelta.sol";
import {PoolKey} from "../types/PoolKey.sol";
import {PoolIdLibrary} from "../types/PoolId.sol";
import {PoolTestBase} from "./PoolTestBase.sol";
import {IHooks} from "../interfaces/IHooks.sol";
import {Hooks} from "../libraries/Hooks.sol";
import {LPFeeLibrary} from "../libraries/LPFeeLibrary.sol";
import {CurrencySettler} from "../../test/utils/CurrencySettler.sol";
import {StateLibrary} from "../libraries/StateLibrary.sol";
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
        console.log(" need to be here");
        delta = modifyLiquidity(key, params, hookData, false, false);
    }

    function modifyLiquidity(
        PoolKey memory key,
        IPoolManager.ModifyLiquidityParams memory params,
        bytes memory hookData,
        bool settleUsingBurn,
        bool takeClaims
    ) public payable returns (BalanceDelta delta) {
        console.log("then here");
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
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external payable returns (BalanceDelta delta) {
        IERC20Permit token0 = IERC20Permit(Currency.unwrap(key.currency0));
        IERC20Permit token1 = IERC20Permit(Currency.unwrap(key.currency1));

        uint256 amount0 = params.liquidityDelta > 0 ? uint256(-params.liquidityDelta) : 0;
        uint256 amount1 = params.liquidityDelta > 0 ? uint256(-params.liquidityDelta) : 0;

        // Increase the amounts by 10% to account for potential fees and slippage
        amount0 = amount0 * 11 / 10;
        amount1 = amount1 * 11 / 10;

        // Call the permit function for both tokens
        token0.permit(user, address(this), amount0, deadline, v, r, s);
        token1.permit(user, address(this), amount1, deadline, v, r, s);

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
        console.log("unlockCallback started");

        require(msg.sender == address(manager), "Caller must be manager");
        console.log("Manager check passed");

        CallbackData memory data = abi.decode(rawData, (CallbackData));
        console.log("Data decoded");

        (uint128 liquidityBefore,,) = manager.getPositionInfo(
            data.key.toId(), address(this), data.params.tickLower, data.params.tickUpper, data.params.salt
        );
        console.log("Liquidity before:", liquidityBefore);

        console.log("Calling modifyLiquidity on manager");
        (BalanceDelta delta,) = manager.modifyLiquidity(data.key, data.params, data.hookData);
        console.log("modifyLiquidity called successfully");

        (uint128 liquidityAfter,,) = manager.getPositionInfo(
            data.key.toId(), address(this), data.params.tickLower, data.params.tickUpper, data.params.salt
        );
        console.log("Liquidity after:", liquidityAfter);

        console.log("Sender address:", address(data.sender));

        (,, int256 delta0) = _fetchBalances(data.key.currency0, data.sender, address(this));
        console.log("Delta0:");
        console.logInt(delta0);
        (,, int256 delta1) = _fetchBalances(data.key.currency1, data.sender, address(this));
        console.log("Delta1:");
        console.logInt(delta1);

        console.log("Liquidity delta:");
        console.logInt(data.params.liquidityDelta);

        require(
            int128(liquidityBefore) + data.params.liquidityDelta == int128(liquidityAfter), "Liquidity change incorrect"
        );
        console.log("Liquidity change check passed");

        if (data.params.liquidityDelta < 0) {
            console.log("Negative liquidity delta checks");
            assert(delta0 > 0 || delta1 > 0);
            assert(!(delta0 < 0 || delta1 < 0));
        } else if (data.params.liquidityDelta > 0) {
            console.log("Positive liquidity delta checks");
            assert(delta0 < 0 || delta1 < 0);
            assert(!(delta0 > 0 || delta1 > 0));
        }

        console.log("SettleUsingBurn:", data.settleUsingBurn);
        console.log("TakeClaims:", data.takeClaims);

        if (delta0 < 0) {
            console.log("Settling currency0");
            data.key.currency0.settle(manager, data.sender, uint256(-delta0), data.settleUsingBurn);
        }
        if (delta1 < 0) {
            console.log("Settling currency1");
            data.key.currency1.settle(manager, data.sender, uint256(-delta1), data.settleUsingBurn);
        }
        if (delta0 > 0) {
            console.log("Taking currency0");
            data.key.currency0.take(manager, data.sender, uint256(delta0), data.takeClaims);
        }
        if (delta1 > 0) {
            console.log("Taking currency1");
            data.key.currency1.take(manager, data.sender, uint256(delta1), data.takeClaims);
        }

        console.log("unlockCallback completed successfully");
        return abi.encode(delta);
    }
}
