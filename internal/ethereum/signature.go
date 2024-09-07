package ethereum

import (
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func VerifySignature(message []byte, signature []byte, address common.Address) bool {
	log.Println("checking")

	// Remove recovery id if present
	if len(signature) == 65 {
		signature = signature[:len(signature)-1]
	}
	log.Println("checking")

	// Add Ethereum prefix
	prefixedMessage := crypto.Keccak256([]byte("\x19Ethereum Signed Message:\n32"), crypto.Keccak256(message))
	log.Println("checking")

	// Recover public key
	pubKey, err := crypto.Ecrecover(prefixedMessage, signature)
	if err != nil {
		return false
	}
	log.Println("checking")

	// Verify recovered address matches expected address
	recoveredAddress := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])
	log.Println("checking", recoveredAddress, address)

	return recoveredAddress == address
}
