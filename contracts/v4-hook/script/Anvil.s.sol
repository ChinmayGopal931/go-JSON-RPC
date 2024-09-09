// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "forge-std/Script.sol";
import {IHooks} from "v4-core/src/interfaces/IHooks.sol";
import {Hooks} from "v4-core/src/libraries/Hooks.sol";
import {PoolManager} from "v4-core/src/PoolManager.sol";
import {IPoolManager} from "v4-core/src/interfaces/IPoolManager.sol";
import {PoolModifyLiquidityTest} from "../test/utils/PoolModifyLiquidityTestPermit.sol";
import {PoolSwapTest} from "../test/utils/PoolSwapTestPermit.sol";
import {PoolDonateTest} from "v4-core/src/test/PoolDonateTest.sol";
import {PoolKey} from "v4-core/src/types/PoolKey.sol";
import {MockERC20} from "solmate/src/test/utils/mocks/MockERC20.sol";
import {Constants} from "v4-core/src/../test/utils/Constants.sol";
import {TickMath} from "v4-core/src/libraries/TickMath.sol";
import {CurrencyLibrary, Currency} from "v4-core/src/types/Currency.sol";
import {Counter} from "../src/Counter.sol";
import {HookMiner} from "../test/utils/HookMiner.sol";
import {MOCKERC20PERMIT} from "../test/utils/MockERC20Permit.sol";
import "@openzeppelin/contracts/token/ERC20/extensions/IERC20Permit.sol";

/// @notice Forge script for deploying v4 & hooks to **anvil**
/// @dev This script only works on an anvil RPC because v4 exceeds bytecode limits
contract CounterScript is Script {
    address constant CREATE2_DEPLOYER = address(0x4e59b44847b379578588920cA78FbF26c0B4956C);
    address public alice;
    uint256 private alicePrivateKey;
    address bob = makeAddr("bob"); // Bob will be our relayer

    function setUp() public {}

    function run() public {
        (alice, alicePrivateKey) = makeAddrAndKey("alice");

        vm.broadcast();
        IPoolManager manager = deployPoolManager();

        // hook contracts must have specific flags encoded in the address
        uint160 permissions = uint160(
            Hooks.BEFORE_SWAP_FLAG | Hooks.AFTER_SWAP_FLAG | Hooks.BEFORE_ADD_LIQUIDITY_FLAG
                | Hooks.BEFORE_REMOVE_LIQUIDITY_FLAG
        );

        // Mine a salt that will produce a hook address with the correct permissions
        (address hookAddress, bytes32 salt) =
            HookMiner.find(CREATE2_DEPLOYER, permissions, type(Counter).creationCode, abi.encode(address(manager)));

        // ----------------------------- //
        // Deploy the hook using CREATE2 //
        // ----------------------------- //
        vm.broadcast();
        Counter counter = new Counter{salt: salt}(manager);
        require(address(counter) == hookAddress, "CounterScript: hook address mismatch");
        console.log("HOOK ADDY:", address(hookAddress));
        console.log("MANAGER ADDY:", address(manager));
        console.log("ALICE ADDY: ", address(alice));
        console.log("BOB ADDY", address(bob));

        // Additional helpers for interacting with the pool
        vm.startBroadcast();
        (PoolModifyLiquidityTest lpRouter, PoolSwapTest swapRouter,) = deployRouters(manager);
        vm.stopBroadcast();

        // test the lifecycle (create pool, add liquidity, swap)
        vm.startBroadcast();
        testLifecycleWithPermit(manager, address(counter), lpRouter, swapRouter);
        vm.stopBroadcast();

        vm.startBroadcast();
        deployAndSeedTestContracts(manager, address(counter), lpRouter, swapRouter);
        vm.stopBroadcast();
    }

    // -----------------------------------------------------------
    // Helpers
    // -----------------------------------------------------------
    function deployPoolManager() internal returns (IPoolManager) {
        return IPoolManager(address(new PoolManager(500000)));
    }

    function deployRouters(IPoolManager manager)
        internal
        returns (PoolModifyLiquidityTest lpRouter, PoolSwapTest swapRouter, PoolDonateTest donateRouter)
    {
        lpRouter = new PoolModifyLiquidityTest(manager);
        swapRouter = new PoolSwapTest(manager);
        donateRouter = new PoolDonateTest(manager);
        console.log("SWAP ROUTER: ", address(swapRouter));
        console.log("lpRouter ROUTER: ", address(lpRouter));
    }

    function deployTokens() internal returns (MOCKERC20PERMIT token0, MOCKERC20PERMIT token1) {
        MOCKERC20PERMIT tokenA = new MOCKERC20PERMIT("MockA", "A", 18);
        MOCKERC20PERMIT tokenB = new MOCKERC20PERMIT("MockB", "B", 18);

        if (uint160(address(tokenA)) < uint160(address(tokenB))) {
            token0 = tokenA;
            token1 = tokenB;
        } else {
            token0 = tokenB;
            token1 = tokenA;
        }

        console.log("TOKEN A", address(token0));
        console.log("TOKEN B", address(token1));
    }

    function testLifecycleWithPermit(
        IPoolManager manager,
        address hook,
        PoolModifyLiquidityTest lpRouter,
        PoolSwapTest swapRouter
    ) internal {
        (MOCKERC20PERMIT token0, MOCKERC20PERMIT token1) = deployTokens();

        // Transfer tokens to Alice
        token0.mint(msg.sender, 100_000 ether);
        token1.mint(msg.sender, 100_000 ether);
        token0.mint(alice, 100_000 ether);
        token1.mint(alice, 100_000 ether);
        token0.mint(address(0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266), 100_000 ether);
        token1.mint(address(0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266), 100_000 ether);

        // bytes memory ZERO_BYTES = new bytes(0);

        // // Initialize the pool
        // int24 tickSpacing = 60;
        // PoolKey memory poolKey =
        //     PoolKey(Currency.wrap(address(token0)), Currency.wrap(address(token1)), 3000, tickSpacing, IHooks(hook));
        // manager.initialize(poolKey, Constants.SQRT_PRICE_1_1, ZERO_BYTES);

        token0.approve(address(lpRouter), type(uint256).max);
        token1.approve(address(lpRouter), type(uint256).max);

        // lpRouter.modifyLiquidity(
        //     poolKey,
        //     IPoolManager.ModifyLiquidityParams(
        //         TickMath.minUsableTick(tickSpacing), TickMath.maxUsableTick(tickSpacing), 100 ether, 0
        //     ),
        //     ZERO_BYTES
        // );

        // // Prepare swap parameters
        // bool zeroForOne = true;
        // int256 amountSpecified = 1 ether;
        // IPoolManager.SwapParams memory params = IPoolManager.SwapParams({
        //     zeroForOne: zeroForOne,
        //     amountSpecified: amountSpecified,
        //     sqrtPriceLimitX96: zeroForOne ? TickMath.MIN_SQRT_PRICE + 1 : TickMath.MAX_SQRT_PRICE - 1
        // });
        // PoolSwapTest.TestSettings memory testSettings =
        //     PoolSwapTest.TestSettings({takeClaims: false, settleUsingBurn: false});

        // // Prepare permit data
        // uint256 deadline = block.timestamp + 3600; // 1 hour from now
        // uint256 value = uint256(amountSpecified) * 11 / 10; // Increase by 10% to account for fees and slippage

        // console.log("VALUE", value);
        // console.log("swap router", address(swapRouter));
        // // vm.startPrank(alice);

        // (uint8 v, bytes32 r, bytes32 s) = generatePermitSignature(
        //     IERC20Permit(address(token0)), alice, address(swapRouter), value, deadline, alicePrivateKey
        // );

        // // vm.stopPrank();

        // console.log("Alice address:", alice);
        // console.log("bob address", address(bob));

        // // Perform the swap with permit (as bob, the relayer)
        // vm.startBroadcast(bob);

        // // Then, perform the swap on behalf of Alice
        // // Call swapWithPermit
        // swapRouter.swapWithPermit(
        //     alice, // user
        //     poolKey,
        //     params,
        //     testSettings,
        //     ZERO_BYTES, // hookData
        //     deadline,
        //     v,
        //     r,
        //     s
        // );

        // vm.stopBroadcast();

        // // Verify the swap results (you may want to add more assertions)
        // console.log("Swap completed successfully");
        // console.log("Token0 balance of Alice:", token0.balanceOf(alice));
        // console.log("Token1 balance of Alice:", token1.balanceOf(alice));
    }

    function deployAndSeedTestContracts(
        IPoolManager manager,
        address hook,
        PoolModifyLiquidityTest lpRouter,
        PoolSwapTest swapRouter
    ) internal {
        MOCKERC20PERMIT tokenC = new MOCKERC20PERMIT("MockC", "C", 18);
        MOCKERC20PERMIT tokenD = new MOCKERC20PERMIT("MockD", "D", 18);

        if (uint160(address(tokenC)) > uint160(address(tokenD))) {
            MOCKERC20PERMIT temp = tokenD;
            tokenD = tokenC;
            tokenC = temp;
        }

        console.log("TOKEN C", address(tokenC));
        console.log("TOKEN D", address(tokenD));

        // Transfer tokens to Alice
        tokenC.mint(msg.sender, 100_000 ether);
        tokenD.mint(msg.sender, 100_000 ether);
        tokenC.mint(alice, 100_000 ether);
        tokenD.mint(alice, 100_000 ether);
        tokenC.mint(address(0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266), 100_000 ether);
        tokenD.mint(address(0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266), 100_000 ether);

        tokenC.approve(address(lpRouter), type(uint256).max);
        tokenD.approve(address(lpRouter), type(uint256).max);

        tokenC.approve(address(swapRouter), type(uint256).max);
        tokenD.approve(address(swapRouter), type(uint256).max);

        bytes memory ZERO_BYTES = new bytes(0);

        // Initialize the pool
        int24 tickSpacing = 60;
        PoolKey memory poolKey =
            PoolKey(Currency.wrap(address(tokenC)), Currency.wrap(address(tokenD)), 3000, tickSpacing, IHooks(hook));

        manager.initialize(poolKey, Constants.SQRT_PRICE_1_1, ZERO_BYTES);

        lpRouter.modifyLiquidity(
            poolKey,
            IPoolManager.ModifyLiquidityParams(
                TickMath.minUsableTick(tickSpacing), TickMath.maxUsableTick(tickSpacing), 100 ether, 0
            ),
            ZERO_BYTES
        );

        // Prepare swap parameters
        bool zeroForOne = true;
        int256 amountSpecified = 1 ether;
        IPoolManager.SwapParams memory params = IPoolManager.SwapParams({
            zeroForOne: zeroForOne,
            amountSpecified: amountSpecified,
            sqrtPriceLimitX96: zeroForOne ? TickMath.MIN_SQRT_PRICE + 1 : TickMath.MAX_SQRT_PRICE - 1
        });

        PoolSwapTest.TestSettings memory testSettings =
            PoolSwapTest.TestSettings({takeClaims: false, settleUsingBurn: false});

        swapRouter.swap(poolKey, params, testSettings, ZERO_BYTES);
    }

    function generatePermitSignature(
        IERC20Permit token,
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        uint256 privateKey
    ) internal view returns (uint8 v, bytes32 r, bytes32 s) {
        bytes32 PERMIT_TYPEHASH =
            keccak256("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)");
        uint256 nonce = token.nonces(owner);

        console.log("Deployment - Owner:", owner);
        console.log("Deployment - Spender:", spender);
        console.log("Deployment - Value:", value);
        console.log("Deployment - Nonce:", nonce);
        console.log("Deployment - Deadline:", deadline);

        bytes32 domainSeparator = token.DOMAIN_SEPARATOR();
        console.log("Deployment - Domain Separator:", uint256(domainSeparator));

        bytes32 permitHash = keccak256(abi.encode(PERMIT_TYPEHASH, owner, spender, value, nonce, deadline));
        console.log("Deployment - Permit Hash:", uint256(permitHash));

        bytes32 digest = keccak256(abi.encodePacked("\x19\x01", domainSeparator, permitHash));
        console.log("Deployment - Final Digest:", uint256(digest));

        (v, r, s) = vm.sign(privateKey, digest);
        console.log("Deployment - v:", uint256(v));
        console.log("Deployment - r:", uint256(r));
        console.log("Deployment - s:", uint256(s));

        address recoveredSigner = ecrecover(digest, v, r, s);
        console.log("Deployment - Recovered signer:", recoveredSigner);

        return (v, r, s);
    }
}
