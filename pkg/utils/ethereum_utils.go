package utils

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"

	"uniswap-v4-rpc/internal/ethereum"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func ApproveTokens(auth *bind.TransactOpts, currency0, currency1 common.Address) error {
	maxApproval := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

	// Get the latest nonce for the account
	nonce, err := ethereum.Client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return fmt.Errorf("failed to retrieve account nonce: %v", err)
	}

	for _, currency := range []common.Address{currency0, currency1} {
		token, err := ethereum.NewERC20(currency)
		if err != nil {
			return fmt.Errorf("failed to instantiate token contract: %v", err)
		}

		for _, router := range []common.Address{ethereum.SwapRouterAddress, ethereum.LPRouterAddress} {
			auth.Nonce = big.NewInt(int64(nonce))
			tx, err := token.Approve(auth, router, maxApproval)
			if err != nil {
				return fmt.Errorf("failed to approve token for router %s: %v", router.Hex(), err)
			}

			receipt, err := bind.WaitMined(context.Background(), ethereum.Client, tx)
			if err != nil {
				return fmt.Errorf("failed to wait for approval transaction to be mined: %v", err)
			}

			if receipt.Status == 0 {
				return fmt.Errorf("approval transaction failed for token %s and router %s", currency.Hex(), router.Hex())
			}

			// Increment the nonce for the next transaction
			nonce++
		}
	}

	return nil
}

func GetBalance(tokenAddress, ownerAddress common.Address) (*big.Int, error) {
	// Check if the tokenAddress is valid
	if tokenAddress == (common.Address{}) {
		return nil, fmt.Errorf("invalid token address: zero address")
	}

	// Check if the ownerAddress is valid
	if ownerAddress == (common.Address{}) {
		return nil, fmt.Errorf("invalid owner address: zero address")
	}

	// Check if there's contract code at the token address
	code, err := ethereum.Client.CodeAt(context.Background(), tokenAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get contract code: %v", err)
	}
	if len(code) == 0 {
		return nil, fmt.Errorf("no contract code at address: %s", tokenAddress.Hex())
	}

	token, err := ethereum.NewERC20(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate token contract: %v", err)
	}

	log.Printf("Token address: %s", tokenAddress.Hex())
	log.Printf("Owner address: %s", ownerAddress.Hex())

	balance, err := token.BalanceOf(&bind.CallOpts{}, ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}

	log.Printf("Balance retrieved: %s", balance.String())

	return balance, nil
}

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		return errors.New("Content-Type is not application/json")
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		return err
	}

	return nil
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func CheckContractDeployment(address common.Address) error {
	code, err := ethereum.Client.CodeAt(context.Background(), address, nil)
	if err != nil {
		return fmt.Errorf("failed to get contract code: %v", err)
	}
	if len(code) == 0 {
		return fmt.Errorf("no contract code at address: %s", address.Hex())
	}
	log.Printf("Contract verified at address: %s", address.Hex())
	return nil
}

func FetchDomainSeparator(tokenAddress common.Address) ([32]byte, error) {
	erc20, err := ethereum.NewERC20(tokenAddress)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to create ERC20 instance: %v", err)
	}

	opts := &bind.CallOpts{Context: context.Background()}
	domainSeparator, err := erc20.DOMAIN_SEPARATOR(opts)
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to fetch domain separator: %v", err)
	}

	return domainSeparator, nil
}

func FetchCurrentNonce(tokenAddress, ownerAddress common.Address) (*big.Int, error) {
	erc20, err := ethereum.NewERC20(tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create ERC20 instance: %v", err)
	}

	opts := &bind.CallOpts{Context: context.Background()}
	nonce, err := erc20.Nonces(opts, ownerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current nonce: %v", err)
	}

	return nonce, nil
}

func GeneratePermitSignature(tokenAddress, owner, spender common.Address, value, deadline *big.Int, alicePrivKey *ecdsa.PrivateKey) (uint8, [32]byte, [32]byte, error) {
	PERMIT_TYPEHASH := crypto.Keccak256([]byte("Swap(address sender,address currency0,address currency1,int256 amountSpecified,bool zeroForOne,uint256 sqrtPriceLimitX96,uint256 nonce)"))
	PERMIT_TYPEHASH = crypto.Keccak256([]byte("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"))

	log.Printf("Permit Typehash: 0x%x", PERMIT_TYPEHASH)

	// Fetch the nonce from the token contract
	nonce, err := FetchCurrentNonce(tokenAddress, owner)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("failed to fetch current nonce: %v", err)
	}
	log.Printf("Nonce : 0x%x", nonce)

	// Fetch the domain separator from the token contract
	domainSeparator, err := FetchDomainSeparator(tokenAddress)
	if err != nil {
		log.Printf("Error getting domain separator: %v", err)
	} else {
		log.Printf("Domain Separator: 0x%x", domainSeparator)
	}
	permitHash := crypto.Keccak256(
		PERMIT_TYPEHASH,
		common.LeftPadBytes(owner.Bytes(), 32),
		common.LeftPadBytes(spender.Bytes(), 32),
		common.LeftPadBytes(value.Bytes(), 32),
		common.LeftPadBytes(nonce.Bytes(), 32),
		common.LeftPadBytes(deadline.Bytes(), 32),
	)

	digest := crypto.Keccak256(
		[]byte("\x19\x01"),
		domainSeparator[:],
		permitHash,
	)

	// Log the encoded permit data
	encodedPermit := crypto.Keccak256(
		PERMIT_TYPEHASH,
		spender.Bytes(),
		ethereum.SwapRouterAddress.Bytes(),
		common.LeftPadBytes(value.Bytes(), 32),
		common.LeftPadBytes(nonce.Bytes(), 32),
		common.LeftPadBytes(deadline.Bytes(), 32),
	)
	log.Printf("Encoded Permit: 0x%x", encodedPermit)

	log.Printf("Digest to Sign: 0x%x", digest)

	signature, err := crypto.Sign(digest, alicePrivKey)
	if err != nil {
		log.Printf("Error signing digest: %v", err)
	}
	log.Printf("Raw Signature: 0x%x", signature)

	v := signature[64] + 27
	r := common.BytesToHash(signature[:32])
	s := common.BytesToHash(signature[32:64])

	log.Printf("v: %d", v)
	log.Printf("r: 0x%x", r)
	log.Printf("s: 0x%x", s)

	// Recover signer address
	pubKey, err := crypto.SigToPub(digest, signature)
	if err != nil {
		log.Printf("Error recovering public key: %v", err)
	} else {
		recoveredAddr := crypto.PubkeyToAddress(*pubKey)
		log.Printf("Recovered signer address: %s", recoveredAddr.Hex())

	}
	publicKey := alicePrivKey.Public()

	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	// Log the address derived from the private key
	derivedAddr := crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Printf("Address derived from PrivateKey: %s", derivedAddr.Hex())

	return v, r, s, nil
}

func MakeAddrAndKey(seed string) (common.Address, *ecdsa.PrivateKey) {
	// Create a deterministic hash from the seed
	hash := crypto.Keccak256([]byte(seed))

	// Create a private key from the hash
	privateKey, err := crypto.ToECDSA(hash[:32]) // Use only the first 32 bytes for the private key
	if err != nil {
		log.Fatalf("Failed to create private key: %v", err)
	}

	// Get the public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to get public key")
	}

	// Generate the Ethereum address from the public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return address, privateKey
}
