package utils

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/internal/ethapi"
)

func Sleep(second int64) {
	fmt.Println(fmt.Sprintf("Sleep %d second", second))
	time.Sleep(time.Duration(second) * time.Second)
}

func GetTransactor(ethClient *ethclient.Client, keyStore *keystore.KeyStore, account accounts.Account, value *big.Int) *bind.TransactOpts {
	nonce, _ := ethClient.PendingNonceAt(context.Background(), account.Address)
	txOpts, _ := bind.NewKeyStoreTransactor(keyStore, account)
	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.Value = value
	txOpts.GasLimit = DefaultGasLimit
	txOpts.GasPrice = big.NewInt(DefaultGasPrice)
	return txOpts
}

func GetCallOpts() *bind.CallOpts {
	callOpts := &bind.CallOpts{
		Pending: true,
		Context: context.Background(),
	}
	return callOpts
}

func DeployBEP20Contract(ethClient *ethclient.Client, wallet *keystore.KeyStore, account accounts.Account, contractData hexutil.Bytes, chainId *big.Int) (common.Hash, error) {
	gasLimit := hexutil.Uint64(DefaultGasLimit)
	nonce, err := ethClient.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return common.Hash{}, err
	}
	gasPrice := hexutil.Big(*big.NewInt(DefaultGasPrice))
	nonceUint64 := hexutil.Uint64(nonce)
	sendTxArgs := &ethapi.SendTxArgs{
		From:     account.Address,
		Data:     &contractData,
		Gas:      &gasLimit,
		GasPrice: &gasPrice,
		Nonce:    &nonceUint64,
	}
	tx := toTransaction(sendTxArgs)

	signTx, err := wallet.SignTx(account, tx, chainId)
	if err != nil {
		return common.Hash{}, err
	}

	return signTx.Hash(), ethClient.SendTransaction(context.Background(), signTx)
}

func SendBNBToTempAccount(rpcClient *ethclient.Client, wallet accounts.Wallet, account accounts.Account, recipient common.Address, amount *big.Int, chainId *big.Int) error {
	gasLimit := hexutil.Uint64(DefaultGasLimit)
	nonce, err := rpcClient.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return err
	}
	gasPrice := hexutil.Big(*big.NewInt(DefaultGasPrice))
	amountBig := hexutil.Big(*amount)
	nonceUint64 := hexutil.Uint64(nonce)
	sendTxArgs := &ethapi.SendTxArgs{
		From:     account.Address,
		To:       &recipient,
		Gas:      &gasLimit,
		GasPrice: &gasPrice,
		Value:    &amountBig,
		Nonce:    &nonceUint64,
	}
	tx := toTransaction(sendTxArgs)

	signTx, err := wallet.SignTx(account, tx, chainId)
	if err != nil {
		return err
	}
	return rpcClient.SendTransaction(context.Background(), signTx)
}

func SendAllRestBNB(ethClient *ethclient.Client, wallet *keystore.KeyStore, account accounts.Account, recipient common.Address, chainId *big.Int) (common.Hash, error) {
	restBalance, _ := ethClient.BalanceAt(context.Background(), account.Address, nil)
	txFee := big.NewInt(1).Mul(big.NewInt(21000), big.NewInt(DefaultGasPrice))
	amount := big.NewInt(1).Sub(restBalance, txFee)
	fmt.Println(fmt.Sprintf("rest balance %s, transfer BNB tx fee %s, transfer %s back to %s", restBalance.String(), txFee.String(), amount.String(), recipient.String()))
	gasLimit := hexutil.Uint64(21000)
	nonce, err := ethClient.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return common.Hash{}, err
	}
	gasPrice := hexutil.Big(*big.NewInt(DefaultGasPrice))
	amountBig := hexutil.Big(*amount)
	nonceUint64 := hexutil.Uint64(nonce)
	sendTxArgs := &ethapi.SendTxArgs{
		From:     account.Address,
		To:       &recipient,
		Gas:      &gasLimit,
		GasPrice: &gasPrice,
		Value:    &amountBig,
		Nonce:    &nonceUint64,
	}
	tx := toTransaction(sendTxArgs)

	signTx, err := wallet.SignTx(account, tx, chainId)
	if err != nil {
		return common.Hash{}, err
	}
	return signTx.Hash(), ethClient.SendTransaction(context.Background(), signTx)
}

func toTransaction(args *ethapi.SendTxArgs) *types.Transaction {
	var input []byte
	if args.Input != nil {
		input = *args.Input
	} else if args.Data != nil {
		input = *args.Data
	}
	if args.To == nil {
		return types.NewContractCreation(uint64(*args.Nonce), (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
	}
	return types.NewTransaction(uint64(*args.Nonce), *args.To, (*big.Int)(args.Value), uint64(*args.Gas), (*big.Int)(args.GasPrice), input)
}

func PrintTxExplorerUrl(msg, txHash string, chainID *big.Int) {
	if chainID.Cmp(big.NewInt(MainnetChainID)) == 0 {
		fmt.Println(fmt.Sprintf(MainnetExplorerTxUrl, msg, txHash))
	} else {
		fmt.Println(fmt.Sprintf(TestnetExplorerTxUrl, msg, txHash))
	}
}

func PrintAddrExplorerUrl(msg, address string, chainID *big.Int) {
	if chainID.Cmp(big.NewInt(MainnetChainID)) == 0 {
		fmt.Println(fmt.Sprintf(MainnetExplorerAddressUrl, msg, address))
	} else {
		fmt.Println(fmt.Sprintf(TestnetExplorerAddressUrl, msg, address))
	}
}

func SendTransactionFromLedger(rpcClient *ethclient.Client, wallet accounts.Wallet, account accounts.Account, recipient common.Address, value *big.Int, data *hexutil.Bytes, chainId *big.Int) (*types.Transaction, error) {
	gasLimit := hexutil.Uint64(DefaultGasLimit)
	nonce, err := rpcClient.PendingNonceAt(context.Background(), account.Address)
	if err != nil {
		return nil, err
	}
	gasPrice := hexutil.Big(*big.NewInt(DefaultGasPrice))
	valueBig := hexutil.Big(*value)
	nonceUint64 := hexutil.Uint64(nonce)
	sendTxArgs := &ethapi.SendTxArgs{
		From:     account.Address,
		To:       &recipient,
		Data:     data,
		Gas:      &gasLimit,
		GasPrice: &gasPrice,
		Value:    &valueBig,
		Nonce:    &nonceUint64,
	}
	tx := toTransaction(sendTxArgs)

	signTx, err := wallet.SignTx(account, tx, chainId)
	if err != nil {
		return nil, err
	}
	return signTx, rpcClient.SendTransaction(context.Background(), signTx)
}