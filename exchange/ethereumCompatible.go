package exchange

import (
	"NewListingBot/config"
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"log"
	"math/big"
	"strings"
)

type EthereumCompatibleInstance interface {
	Buy(tokenABI string, ownerAddress string, contractAddress string) (string, error)
	Withdraw(tokenABI string, ownerAddress string, contractAddress string) (string, error)
}

type EthereumCompatible struct {
	infuraURL    string
	privateKey   string
	amountInWei  *big.Int
	contractAddr common.Address
	cfg          config.Config
	ChainID      int
}

func NewEthereumExchange() EthereumCompatibleInstance {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	return &EthereumCompatible{
		infuraURL:  cfg.EthereumInfuraURL,
		privateKey: cfg.EthereumPrivateKey,
		ChainID:    cfg.EthereumChainID,
	}
}
func NewBinanceExchange() *EthereumCompatible {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	return &EthereumCompatible{
		infuraURL:  cfg.BinanceInfuraURL,
		privateKey: cfg.BinancePrivateKey,
		ChainID:    cfg.BinanceChainID,
	}
}
func NewPolygonExchange() *EthereumCompatible {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	return &EthereumCompatible{
		infuraURL:  cfg.PolygonInfuraURL,
		privateKey: cfg.PolygonPrivateKey,
		ChainID:    cfg.PolygonChainID,
	}
}
func NewSepoliaExchange() *EthereumCompatible {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	return &EthereumCompatible{
		infuraURL:  cfg.SepoliaInfuraURL,
		privateKey: cfg.SepoliaPrivateKey,
		ChainID:    cfg.SepoliaChainID,
	}
}

func (e *EthereumCompatible) Buy(tokenABI string, ownerAddress string, contractAddress string) (string, error) {
	rpcClient, err := rpc.Dial(e.infuraURL)
	if err != nil {
		return "", err
	}
	client := ethclient.NewClient(rpcClient)

	// Set up a client to interact with the Ethereum network
	tokenInstance, err := setupTokenInstance(tokenABI)
	if err != nil {
		return "", err
	}

	// Construct the data field
	tokenAmount := big.NewInt(1000000000000000000) // buy 1 token, for example
	input, err := tokenInstance.Pack("transfer", common.HexToAddress(ownerAddress), tokenAmount)
	if err != nil {
		return "", err
	}

	// Prepare the transaction
	txData, err := prepareTransaction(e.privateKey, client, common.HexToAddress(contractAddress), input)
	if err != nil {
		return "", err
	}

	// Increase the gas limit
	// Sign and broadcast the transaction
	txHash, err := signAndBroadcastTransaction(client, e.privateKey, common.HexToAddress(contractAddress), txData, input, e.ChainID)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func (e *EthereumCompatible) Withdraw(tokenABI string, ownerAddress string, contractAddress string) (string, error) {
	rpcClient, err := rpc.Dial(e.infuraURL)
	if err != nil {
		return "", err
	}
	client := ethclient.NewClient(rpcClient)

	// Set up a client to interact with the Ethereum network
	tokenInstance, err := setupTokenInstance(tokenABI)
	if err != nil {
		return "", err
	}

	// Construct the data field
	tokenAmount := big.NewInt(1000000000000000000) // buy 1 token, for example

	input, err := tokenInstance.Pack("withdraw", common.HexToAddress(ownerAddress), tokenAmount) // withdraw 1 token, for example
	if err != nil {
		return "", err
	}

	// Prepare the transaction
	txData, err := prepareTransaction(e.privateKey, client, common.HexToAddress(ownerAddress), input)
	if err != nil {
		return "", err
	}

	// Sign and broadcast the transaction
	txHash, err := signAndBroadcastTransaction(client, e.privateKey, common.HexToAddress(contractAddress), txData, input, e.ChainID)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

func setupTokenInstance(tokenABI string) (*abi.ABI, error) {

	// Get the ABI for the token contract
	contractABI, err := abi.JSON(strings.NewReader(tokenABI)) // Replace TokenABI with the actual ABI
	if err != nil {
		return nil, err
	}

	return &contractABI, nil
}

func prepareTransaction(privateKeyStr string, client *ethclient.Client, contractAddress common.Address, input []byte) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	value := big.NewInt(0)     // in wei (1 eth = 10^18 wei)
	gasLimit := uint64(200000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	// Prepare the transaction by setting the from address, value, gas limit, and gas price
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &contractAddress,
		Value:    value,
		Data:     input,
	})

	return tx, nil
}

func signAndBroadcastTransaction(client *ethclient.Client, privateKeyStr string, contractAddr common.Address, tx *types.Transaction, input []byte, chainID int) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyStr)
	if err != nil {
		return "", err
	}

	// Create a signer
	signer := types.NewEIP155Signer(big.NewInt(int64(chainID))) // Replace 1 with your chain ID

	// Sign the transaction
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", err
	}

	// Broadcast the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}
