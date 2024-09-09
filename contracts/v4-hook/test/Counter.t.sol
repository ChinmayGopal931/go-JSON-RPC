// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "forge-std/Script.sol";

import {IHooks} from "v4-core/src/interfaces/IHooks.sol";
import {Hooks} from "v4-core/src/libraries/Hooks.sol";
import {TickMath} from "v4-core/src/libraries/TickMath.sol";
import {IPoolManager} from "v4-core/src/interfaces/IPoolManager.sol";
import {PoolKey} from "v4-core/src/types/PoolKey.sol";
import {BalanceDelta} from "v4-core/src/types/BalanceDelta.sol";
import {PoolId, PoolIdLibrary} from "v4-core/src/types/PoolId.sol";
import {CurrencyLibrary, Currency} from "v4-core/src/types/Currency.sol";
import {PoolSwapTest} from "./utils/PoolSwapTestPermit.sol";

import {Counter} from "../src/Counter.sol";
import {StateLibrary} from "v4-core/src/libraries/StateLibrary.sol";
import {PositionConfig} from "v4-periphery/src/libraries/PositionConfig.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Permit.sol";
import {MOCKERC20PERMIT} from "./utils/MockERC20Permit.sol";
import {MockERC20} from "solmate/src/test/utils/mocks/MockERC20.sol";

import {IPositionManager} from "v4-periphery/src/interfaces/IPositionManager.sol";
import {EasyPosm} from "./utils/EasyPosm.sol";
import {Fixtures} from "./utils/Fixtures.sol";

contract CounterTest is Test, Fixtures {
    using EasyPosm for IPositionManager;
    using PoolIdLibrary for PoolKey;
    using CurrencyLibrary for Currency;
    using StateLibrary for IPoolManager;

    Counter hook;
    PoolId poolId;

    uint256 tokenId;
    PositionConfig config;

    function setUp() public {
        // creates the pool manager, utility routers, and test tokens
        deployFreshManagerAndRouters();
        deployMintAndApprove2Currencies();

        deployAndApprovePosm(manager);

        // Deploy the hook to an address with the correct flags
        address flags = address(
            uint160(
                Hooks.BEFORE_SWAP_FLAG | Hooks.AFTER_SWAP_FLAG | Hooks.BEFORE_ADD_LIQUIDITY_FLAG
                    | Hooks.BEFORE_REMOVE_LIQUIDITY_FLAG
            ) ^ (0x4444 << 144) // Namespace the hook to avoid collisions
        );
        bytes memory constructorArgs = abi.encode(manager); //Add all the necessary constructor arguments from the hook
        deployCodeTo("Counter.sol:Counter", constructorArgs, flags);
        hook = Counter(flags);

        // Create the pool
        key = PoolKey(currency0, currency1, 3000, 60, IHooks(hook));
        poolId = key.toId();
        manager.initialize(key, SQRT_PRICE_1_1, ZERO_BYTES);

        // Provide full-range liquidity to the pool
        config = PositionConfig({
            poolKey: key,
            tickLower: TickMath.minUsableTick(key.tickSpacing),
            tickUpper: TickMath.maxUsableTick(key.tickSpacing)
        });
        (tokenId,) = posm.mint(
            config,
            10_000e18,
            MAX_SLIPPAGE_ADD_LIQUIDITY,
            MAX_SLIPPAGE_ADD_LIQUIDITY,
            address(this),
            block.timestamp,
            ZERO_BYTES
        );
    }

    function testSwapWithInsufficientFunds() public {
        // Setup
        address poorUser = makeAddr("poorUser");
        uint256 initialBalance = 0.1 ether; // A small balance, not enough for the swap

        MockERC20 token0 = MockERC20(Currency.unwrap(currency0));
        MockERC20 token1 = MockERC20(Currency.unwrap(currency1));

        // Give the user a small balance
        token0.mint(poorUser, initialBalance);
        token1.mint(poorUser, initialBalance);

        // Approve the swapRouter to spend tokens on behalf of poorUser
        vm.startPrank(poorUser);
        token0.approve(address(swapRouter), type(uint256).max);
        token1.approve(address(swapRouter), type(uint256).max);
        vm.stopPrank();

        // Set up swap parameters
        bool zeroForOne = true;
        int256 amountSpecified = 1 ether; // More than the user's balance

        // Attempt the swap as poorUser
        vm.prank(poorUser);
        vm.expectRevert();
        swap(key, zeroForOne, -amountSpecified, ZERO_BYTES);

        // Verify balances haven't changed
        assertEq(token0.balanceOf(poorUser), initialBalance);
        assertEq(token1.balanceOf(poorUser), initialBalance);
    }

    function testSwapRevert() public {
        address bob = makeAddr("bob"); // Bob will be our relayer
        (address alice, uint256 aliceKey) = makeAddrAndKey("alice");

        // Mint tokens to Alice
        MockERC20 token0 = MockERC20(Currency.unwrap(currency0));
        MockERC20 token1 = MockERC20(Currency.unwrap(currency1));
        token0.mint(alice, 100_000 ether);
        token1.mint(alice, 100_000 ether);

        // Perform a test swap
        bool zeroForOne = true;
        int256 amountSpecified = 1e18;
        uint256 deadline = block.timestamp + 3600; // 1 hour from now
        uint256 value = uint256(amountSpecified) * 11 / 10;

        // Generate permit signature for Alice
        (uint8 v, bytes32 r, bytes32 s) = generatePermitSignature(
            ((MockERC20(Currency.unwrap(currency0)))), alice, address(swapRouter), uint256(value), deadline, aliceKey
        );

        vm.prank(alice);
        token0.approve(address(swapRouter), 0);
        vm.prank(alice);
        token1.approve(address(swapRouter), 0);

        vm.prank(bob);
        token0.approve(address(swapRouter), 10e18);
        vm.prank(bob);
        token1.approve(address(swapRouter), 10e18);

        console.log("alice", token0.allowance(alice, address(swapRouter)), address(alice));
        console.log("bob ", token0.allowance(bob, address(swapRouter)), address(bob));

        // Check initial allowance alice (should be 0)
        assertEq(token0.allowance(alice, address(swapRouter)), 0);

        bool isNativeInput = zeroForOne && key.currency0.isNative();
        if (isNativeInput) require(0 > amountSpecified, "Use swapNativeInput() for native-token exact-output swaps");

        uint256 aliceBalanceBefore = MockERC20(Currency.unwrap(currency0)).balanceOf(alice);
        console.log("Alice balance before swap:", aliceBalanceBefore);

        // Expect revert when trying to swap with no allowance
        vm.expectRevert();
        vm.prank(bob);
        BalanceDelta swapDelta = swapRouter.swap(
            key,
            IPoolManager.SwapParams({
                zeroForOne: zeroForOne,
                amountSpecified: int256(amountSpecified),
                sqrtPriceLimitX96: zeroForOne ? MIN_PRICE_LIMIT : MAX_PRICE_LIMIT
            }),
            PoolSwapTest.TestSettings({takeClaims: false, settleUsingBurn: false}),
            ZERO_BYTES
        );

        vm.prank(alice);
        token0.approve(address(swapRouter), 10e18);

        //Ensure swap works after approval is manually given
        vm.prank(alice);
        swapDelta = swapRouter.swap(
            key,
            IPoolManager.SwapParams({
                zeroForOne: zeroForOne,
                amountSpecified: int256(amountSpecified),
                sqrtPriceLimitX96: zeroForOne ? MIN_PRICE_LIMIT : MAX_PRICE_LIMIT
            }),
            PoolSwapTest.TestSettings({takeClaims: false, settleUsingBurn: false}),
            ZERO_BYTES
        );

        uint256 aliceBalanceAfter = MockERC20(Currency.unwrap(currency0)).balanceOf(alice);

        console.log("balance before and after", aliceBalanceBefore, aliceBalanceAfter);

        // Verify hook counts
        assertEq(hook.beforeSwapCount(poolId), 1);
        assertEq(hook.afterSwapCount(poolId), 1);
    }

    function testPermitSignature() public {
        address bob = makeAddr("bob"); // Bob will be our relayer
        (address alice, uint256 aliceKey) = makeAddrAndKey("alice");

        // Mint tokens to Alice
        MockERC20 token0 = MockERC20(Currency.unwrap(currency0));
        MockERC20 token1 = MockERC20(Currency.unwrap(currency1));
        token0.mint(alice, 100_000 ether);
        token1.mint(alice, 100_000 ether);

        // Perform a test swap
        bool zeroForOne = true;
        int256 amountSpecified = 1e18;
        uint256 deadline = block.timestamp + 3600; // 1 hour from now
        uint256 value = uint256(amountSpecified) * 11 / 10;

        // Generate permit signature for Alice
        (uint8 v, bytes32 r, bytes32 s) = generatePermitSignature(
            ((MockERC20(Currency.unwrap(currency0)))), alice, address(swapRouter), uint256(value), deadline, aliceKey
        );

        vm.prank(alice);
        token0.approve(address(swapRouter), 0);
        vm.prank(alice);
        token1.approve(address(swapRouter), 0);

        vm.prank(bob);
        token0.approve(address(swapRouter), 10e18);
        vm.prank(bob);
        token1.approve(address(swapRouter), 10e18);

        console.log("alice", token0.allowance(alice, address(swapRouter)), address(alice));
        console.log("bob ", token0.allowance(bob, address(swapRouter)), address(bob));

        // Check initial allowance alice (should be 0)
        assertEq(token0.allowance(alice, address(swapRouter)), 0);

        bool isNativeInput = zeroForOne && key.currency0.isNative();
        if (isNativeInput) require(0 > amountSpecified, "Use swapNativeInput() for native-token exact-output swaps");

        uint256 aliceBalanceBefore = MockERC20(Currency.unwrap(currency0)).balanceOf(alice);
        console.log("Alice balance before swap:", aliceBalanceBefore);

        vm.prank(bob);
        BalanceDelta swapDelta = swapRouter.swapWithPermit(
            alice,
            key,
            IPoolManager.SwapParams({
                zeroForOne: zeroForOne,
                amountSpecified: int256(amountSpecified),
                sqrtPriceLimitX96: zeroForOne ? MIN_PRICE_LIMIT : MAX_PRICE_LIMIT
            }),
            PoolSwapTest.TestSettings({takeClaims: false, settleUsingBurn: false}),
            ZERO_BYTES,
            deadline,
            v,
            r,
            s
        );

        uint256 aliceBalanceAfter = MockERC20(Currency.unwrap(currency0)).balanceOf(alice);
        console.log("Alice balance after swap:", aliceBalanceAfter);

        // Verify hook counts
        assertEq(hook.beforeSwapCount(poolId), 1);
        assertEq(hook.afterSwapCount(poolId), 1);
    }

    function testCounterHooks() public {
        // positions were created in setup()
        assertEq(hook.beforeAddLiquidityCount(poolId), 1);
        assertEq(hook.beforeRemoveLiquidityCount(poolId), 0);

        assertEq(hook.beforeSwapCount(poolId), 0);
        assertEq(hook.afterSwapCount(poolId), 0);

        // Perform a test swap //
        bool zeroForOne = true;
        int256 amountSpecified = -1e18; // negative number indicates exact input swap!
        BalanceDelta swapDelta = swap(key, zeroForOne, amountSpecified, ZERO_BYTES);
        // ------------------- //

        assertEq(int256(swapDelta.amount0()), amountSpecified);

        assertEq(hook.beforeSwapCount(poolId), 1);
        assertEq(hook.afterSwapCount(poolId), 1);
    }

    function testLiquidityHooks() public {
        // positions were created in setup()
        assertEq(hook.beforeAddLiquidityCount(poolId), 1);
        assertEq(hook.beforeRemoveLiquidityCount(poolId), 0);

        // remove liquidity
        uint256 liquidityToRemove = 1e18;
        posm.decreaseLiquidity(
            tokenId,
            config,
            liquidityToRemove,
            MAX_SLIPPAGE_REMOVE_LIQUIDITY,
            MAX_SLIPPAGE_REMOVE_LIQUIDITY,
            address(this),
            block.timestamp,
            ZERO_BYTES
        );

        assertEq(hook.beforeAddLiquidityCount(poolId), 1);
        assertEq(hook.beforeRemoveLiquidityCount(poolId), 1);
    }

    function testModifyLiquidityWithPermit() public {
        address bob = makeAddr("bob"); // Bob will be our relayer
        (address alice, uint256 aliceKey) = makeAddrAndKey("alice");

        // Mint tokens to Alice
        MockERC20 token0 = MockERC20(Currency.unwrap(currency0));
        MockERC20 token1 = MockERC20(Currency.unwrap(currency1));
        token0.mint(alice, 100_000 ether);
        token1.mint(alice, 100_000 ether);

        // Prepare modifyLiquidity parameters
        IPoolManager.ModifyLiquidityParams memory params = IPoolManager.ModifyLiquidityParams({
            tickLower: TickMath.minUsableTick(key.tickSpacing),
            tickUpper: TickMath.maxUsableTick(key.tickSpacing),
            liquidityDelta: 1e18,
            salt: 0
        });

        uint256 deadline = block.timestamp + 3600; // 1 hour from now
        uint256 value = uint256((params.liquidityDelta)); // Convert to uint128 first to avoid overflow

        // Generate permit signatures for Alice
        (uint8 v0, bytes32 r0, bytes32 s0) =
            generatePermitSignature(token0, alice, address(modifyLiquidityRouter), value, deadline, aliceKey);
        (uint8 v1, bytes32 r1, bytes32 s1) =
            generatePermitSignature(token1, alice, address(modifyLiquidityRouter), value, deadline, aliceKey);

        // Remove any existing approvals
        vm.prank(alice);
        token0.approve(address(modifyLiquidityRouter), 0);
        vm.prank(alice);
        token1.approve(address(modifyLiquidityRouter), 0);

        // Check initial allowances (should be 0)
        assertEq(token0.allowance(alice, address(modifyLiquidityRouter)), 0);
        assertEq(token1.allowance(alice, address(modifyLiquidityRouter)), 0);

        uint256 aliceBalance0Before = token0.balanceOf(alice);
        uint256 aliceBalance1Before = token1.balanceOf(alice);

        // Perform modifyLiquidityWithPermit as Bob (relayer)
        vm.prank(bob);
        BalanceDelta delta = modifyLiquidityRouter.modifyLiquidityWithPermit(
            alice, key, params, ZERO_BYTES, false, false, deadline, v0, r0, s0, v1, r1, s1
        );

        uint256 aliceBalance0After = token0.balanceOf(alice);
        uint256 aliceBalance1After = token1.balanceOf(alice);

        // Verify balances have changed
        assertTrue(aliceBalance0After < aliceBalance0Before, "Token0 balance should have decreased");
        assertTrue(aliceBalance1After < aliceBalance1Before, "Token1 balance should have decreased");

        // Verify liquidity was added
        (uint128 liquidity,,) = manager.getPositionInfo(
            key.toId(), address(modifyLiquidityRouter), params.tickLower, params.tickUpper, params.salt
        );
    }

    function testAddLiquidityNormal() public {
        // Setup
        address alice = makeAddr("alice");
        MockERC20 token0 = MockERC20(Currency.unwrap(currency0));
        MockERC20 token1 = MockERC20(Currency.unwrap(currency1));

        // Mint tokens to Alice
        uint256 mintAmount = 100_000 ether;
        token0.mint(alice, mintAmount);
        token1.mint(alice, mintAmount);

        // Approve tokens for PoolModifyLiquidityTest
        vm.startPrank(alice);
        token0.approve(address(modifyLiquidityRouter), type(uint256).max);
        token1.approve(address(modifyLiquidityRouter), type(uint256).max);
        vm.stopPrank();

        // Prepare modifyLiquidity parameters
        IPoolManager.ModifyLiquidityParams memory params = IPoolManager.ModifyLiquidityParams({
            tickLower: TickMath.minUsableTick(key.tickSpacing),
            tickUpper: TickMath.maxUsableTick(key.tickSpacing),
            liquidityDelta: 1e18,
            salt: 0
        });

        // Record initial balances
        uint256 aliceBalance0Before = token0.balanceOf(alice);
        uint256 aliceBalance1Before = token1.balanceOf(alice);

        // Add liquidity
        vm.prank(alice);
        BalanceDelta delta = modifyLiquidityRouter.modifyLiquidity(key, params, ZERO_BYTES);

        // Verify balances have changed
        uint256 aliceBalance0After = token0.balanceOf(alice);
        uint256 aliceBalance1After = token1.balanceOf(alice);
        assert(aliceBalance0After < aliceBalance0Before);
        assert(aliceBalance1After < aliceBalance1Before);

        // Verify liquidity was added
        (uint128 liquidity,,) = manager.getPositionInfo(
            key.toId(), address(modifyLiquidityRouter), params.tickLower, params.tickUpper, params.salt
        );

        // Verify hook counts (if applicable)
        assertEq(hook.beforeAddLiquidityCount(poolId), 2);

        // Verify BalanceDelta
        assert(delta.amount0() < 0);
        assert(delta.amount1() < 0);
    }

    //HELPER

    function generatePermitSignature(
        MockERC20 token,
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        uint256 privateKey
    ) public view returns (uint8 v, bytes32 r, bytes32 s) {
        bytes32 PERMIT_TYPEHASH =
            keccak256("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)");
        uint256 nonce = token.nonces(owner);

        console.log("Deployment - Owner:", owner);
        console.log("Deployment - Spender:", spender);
        console.log("Deployment - Value:", value);
        console.log("Deployment - Nonce:", nonce);
        console.log("Deployment - Deadline:", deadline);

        bytes32 domainSeparator = token.DOMAIN_SEPARATOR();

        bytes32 permitHash = keccak256(abi.encode(PERMIT_TYPEHASH, owner, spender, value, nonce, deadline));

        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, permitHash));

        (v, r, s) = vm.sign(privateKey, digest);

        address recoveredSigner = ecrecover(digest, v, r, s);

        console.log("v:", v);
        console.log("r:", uint256(r));
        console.log("s:", uint256(s));

        console.log("Debug Permit:");
        console.log("  Domain Separator:", uint256(domainSeparator));
        console.log("  Nonce:", nonce);
        console.log("  Permit Hash:", uint256(permitHash));
        console.log("  Digest:", uint256(digest));
        console.log("  Recovered Signer:", recoveredSigner);
        console.log("  Expected Signer:", owner);
        console.log("Deployment - Recovered signer:", recoveredSigner);

        return (v, r, s);
    }
}
