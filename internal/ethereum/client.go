package ethereum

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	Client     *ethclient.Client
	PrivateKey *ecdsa.PrivateKey
)

func InitClient(nodeURL string) error {
	var err error
	Client, err = ethclient.Dial(nodeURL)
	if err != nil {
		return err
	}
	return nil
}

func SetPrivateKey(hexPrivateKey string) error {
	var err error
	PrivateKey, err = crypto.HexToECDSA(hexPrivateKey)
	if err != nil {
		return err
	}
	return nil
}
