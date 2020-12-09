package svm

import (
	"WuyaChain/common/errors"
	"WuyaChain/core/state"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
	"math/big"
)

// Context for other vm constructs
type Context struct {
	Tx          *types.Transaction
	TxIndex     int
	Statedb     *state.Statedb
	BlockHeader *types.BlockHeader
	BcStore     store.BlockchainStore
}

// Process the tx
func Process(ctx *Context, height uint64) (*types.Receipt, error) {
	// check the tx against the latest statedb, e.g. balance, nonce.
	if err := ctx.Tx.ValidateState(ctx.Statedb, height); err != nil {
		return nil, errors.NewStackedError(err, "failed to validate tx against statedb")
	}

	// Pay intrinsic gas all the time
	gasLimit := ctx.Tx.Data.GasLimit
	intrGas := ctx.Tx.IntrinsicGas()
	if gasLimit < intrGas {
		return nil, types.ErrIntrinsicGas
	}
	//leftOverGas := gasLimit - intrGas
	//
	//// init statedb and set snapshot
	//var err error
	var receipt *types.Receipt
	snapshot := ctx.Statedb.Prepare(ctx.TxIndex)

	//// create or execute contract
	//if contract := system.GetContractByAddress(ctx.Tx.Data.To); contract != nil { // system contract
	//	receipt, err = processSystemContract(ctx, contract, snapshot, leftOverGas)
	//} else if ctx.Tx.IsCrossShardTx() && !ctx.Tx.Data.To.IsEVMContract() { // cross shard tx
	//	return processCrossShardTransaction(ctx, snapshot)
	//} else { // evm
	//	receipt, err = processEvmContract(ctx, leftOverGas)
	//}
	//// fmt.Println("svm.go-59, receipt.result", receipt.Result)
	//// account balance is not enough (account.balance < tx.amount)
	//if err == vm.ErrInsufficientBalance {
	//	return nil, revertStatedb(ctx.Statedb, snapshot, err)
	//}
	//
	//if err != nil {
	//	if height <= common.SmartContractNonceForkHeight {
	//		// fmt.Println("smart contract OLD logic")
	//		ctx.Statedb.RevertToSnapshot(snapshot)
	//		receipt.Failed = true
	//		receipt.Result = []byte(err.Error())
	//
	//	} else {
	//		// fmt.Println("smart contract NEW logic")
	//		databaseAccountNonce := ctx.Statedb.GetNonce(ctx.Tx.Data.From)
	//		setNonce := databaseAccountNonce
	//		if ctx.Tx.Data.AccountNonce >= databaseAccountNonce {
	//			setNonce = ctx.Tx.Data.AccountNonce + 1
	//		}
	//		ctx.Statedb.RevertToSnapshot(snapshot)
	//		ctx.Statedb.SetNonce(ctx.Tx.Data.From, setNonce)
	//		receipt.Failed = true
	//		receipt.Result = []byte(err.Error())
	//	}
	//
	//}

	// include the intrinsic gas
	receipt.UsedGas += intrGas

	// refund gas, capped to half of the used gas.
	refund := ctx.Statedb.GetRefund()
	if maxRefund := receipt.UsedGas / 2; refund > maxRefund {
		refund = maxRefund
	}
	receipt.UsedGas -= refund

	return handleFee(ctx, receipt, snapshot)
}


func handleFee(ctx *Context, receipt *types.Receipt, snapshot int) (*types.Receipt, error) {
	// Calculating the total fee
	// For normal tx: fee = 20k * 1 Fan/gas = 0.0002 Seele
	// For contract tx, average gas per tx is about 100k on ETH, fee = 100k * 1Fan/gas = 0.001 Seele
	usedGas := new(big.Int).SetUint64(receipt.UsedGas)
	totalFee := new(big.Int).Mul(usedGas, ctx.Tx.Data.GasPrice)

	// Transfer fee to coinbase
	// Note, the sender should always have enough balance.
	ctx.Statedb.SubBalance(ctx.Tx.Data.From, totalFee)
	ctx.Statedb.AddBalance(ctx.BlockHeader.Creator, totalFee)
	receipt.TotalFee = totalFee.Uint64()

	// Record statedb hash
	var err error
	if receipt.PostState, err = ctx.Statedb.Hash(); err != nil {
		err = errors.NewStackedError(err, "failed to get statedb root hash")
		return nil, revertStatedb(ctx.Statedb, snapshot, err)
	}

	// Add logs
	receipt.Logs = ctx.Statedb.GetCurrentLogs()
	if receipt.Logs == nil {
		receipt.Logs = make([]*types.Log, 0)
	}

	return receipt, nil
}

func revertStatedb(statedb *state.Statedb, snapshot int, err error) error {
	statedb.RevertToSnapshot(snapshot)
	return err
}
