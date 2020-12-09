package types

import (
	"errors"
	"WuyaChain/common"
	"WuyaChain/crypto"
	"WuyaChain/trie"
	"fmt"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
)

// TxType represents transaction type
type TxType byte

// Transaction types
const (
	TxTypeRegular TxType = iota
	TxTypeReward
)

const (
	defaultMaxPayloadSize = 32 * 1024

	// TransactionPreSize is the transaction size excluding payload size
	TransactionPreSize = 177
)

type indexInBlock struct {
	BlockHash common.Hash
	Index     uint // index in block body
}

// TxIndex represents an index that used to query block info by tx hash.
type TxIndex indexInBlock


var (
	// ErrPriceNil is returned when the transaction gas price is nil.
	ErrPriceNil = errors.New("gas price is nil")
	// ErrAmountNegative is returned when the transaction amount is negative.
	ErrAmountNegative = errors.New("amount is negative")
	// ErrAmountNil is returned when the transaction amount is nil.
	ErrAmountNil = errors.New("amount is null")
	// ErrPriceNegative is returned when the transaction gas price is negative or zero.
	ErrPriceNegative = errors.New("gas price is negative or zero")
	// ErrIntrinsicGas is returned if the tx gas is too low.
	ErrIntrinsicGas = errors.New("intrinsic gas too low")
	// MaxPayloadSize limits the payload size to prevent malicious transactions.
	MaxPayloadSize = defaultMaxPayloadSize
	// ErrPayloadOversized is returned when the payload size is larger than the MaxPayloadSize.
	ErrPayloadOversized = errors.New("oversized payload")

	// ErrPayloadEmpty is returned when create or call a contract without payload.
	ErrPayloadEmpty = errors.New("empty payload")
	// ErrSigMissing is returned when the transaction signature is missing.
	ErrSigMissing = errors.New("signature missing")
	// ErrHashMismatch is returned when the transaction hash and data mismatch.
	ErrHashMismatch = errors.New("hash mismatch")
	// verified tx signature cache <txHash, error>
	sigCache = common.MustNewCache(20 * 1024)
	// ErrSigInvalid is returned when the transaction signature is invalid.
	ErrSigInvalid = errors.New("signature is invalid")


)

type Transaction struct {
	Hash      common.Hash
	Data      TransactionData
	Signature crypto.Signature
}

type TransactionData struct {
	Type         TxType         // Transaction type
	From common.Address
	To common.Address
    Amount *big.Int
	AccountNonce uint64
	GasPrice     *big.Int       // Transaction gas price
	GasLimit     uint64         // Maximum gas for contract creation/execution
	Timestamp    uint64         // Timestamp is used for the miner reward transaction, referring to the block timestamp
	Payload      common.Bytes   // Payload is the extra data of the transaction
}

var (
	emptyTxRootHash = common.EmptyHash
)

type stateDB interface {
	GetBalance(common.Address) *big.Int
	GetNonce(common.Address) uint64
}

// MerkleRootHash calculates and returns the merkle root hash of the specified transactions.
// If the given transactions are empty, return empty hash.
func MerkleRootHash(txs []*Transaction) common.Hash {
	if len(txs) == 0 {
		return emptyTxRootHash
	}

	trie := GetTxTrie(txs)
	return trie.Hash()
}

// GetTxTrie generate trie according the txs
func GetTxTrie(txs []*Transaction) *trie.Trie {
	emptyTrie, err := trie.NewTrie(common.EmptyHash, make([]byte, 0), nil)
	if err != nil {
		panic(err)
	}

	for _, tx := range txs {
		buff := common.SerializePanic(tx)
		if tx.Hash != common.EmptyHash {
			emptyTrie.Put(tx.Hash.Bytes(), buff)
		} else {
			emptyTrie.Put(tx.CalculateHash().Bytes(), buff)
		}
	}

	return emptyTrie
}

// GetTransactionsSize return the transaction size
func GetTransactionsSize(txs []*Transaction) int {
	size := 0
	for _, tx := range txs {
		size += tx.Size()
	}
	return size
}

// Size return the transaction size
func (tx *Transaction) Size() int {
	return TransactionPreSize + len(tx.Data.Payload)
}

// CalculateHash calculates and returns the transaction hash.
func (tx *Transaction) CalculateHash() common.Hash {
	return crypto.MustHash(tx.Data)
}

func (tx *Transaction) FromAccount() common.Address {
	return tx.Data.From
}

func (tx *Transaction) ToAccount() common.Address {
	return tx.Data.To
}

func (tx *Transaction) Nonce() uint64 {
	return tx.Data.AccountNonce
}

func (tx *Transaction) Price() *big.Int {
	return tx.Data.GasPrice
}

func (tx *Transaction) GetHash() common.Hash {
	return tx.Hash
}

// BatchValidateTxs validates the state independent fields of specified txs in multiple threads.
// Because the signature verification is time consuming (see test Benchmark_Transaction_ValidateWithoutState),
// once a block includes too many txs (e.g. 5000), the txs validation will consume too much time.
func BatchValidateTxs(txs []*Transaction) error {
	return BatchValidate(func(index int) error {
		return txs[index].ValidateWithoutState(true, true)
	}, len(txs))
}


func BatchValidate(handler func(index int) error, len int) error {
	threads := runtime.NumCPU() / 2 // in case of CPU 100%

	// single thread for few CPU kernel or few txs to validate.
	if threads <= 1 || len < threads {
		for i := 0; i < len; i++ {
			if err := handler(i); err != nil {
				return err
			}
		}

		return nil
	}

	// parallel validates txs
	var err error
	var hasErr uint32
	wg := sync.WaitGroup{}

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()

			for j := offset; j < len && atomic.LoadUint32(&hasErr) == 0; j += threads {
				if e := handler(j); e != nil {
					if atomic.CompareAndSwapUint32(&hasErr, 0, 1) {
						err = e
					}

					break
				}
			}
		}(i)
	}

	wg.Wait()

	return err
}


// ValidateWithoutState validates state independent fields in tx.
func (tx *Transaction) ValidateWithoutState(signNeeded bool, shardNeeded bool) error {
	// validate from/to address
	if err := tx.Data.From.Validate(); err != nil {
		return err
	}

	if err := tx.Data.To.Validate(); err != nil {
		return err
	}

	// validate amount
	if tx.Data.Amount == nil {
		return ErrAmountNil
	}

	if tx.Data.Amount.Sign() < 0 {
		return ErrAmountNegative
	}

	// validate gas price
	if tx.Data.GasPrice == nil {
		return ErrPriceNil
	}

	if tx.Data.GasPrice.Sign() <= 0 {
		return ErrPriceNegative
	}

	// validate payload
	if len(tx.Data.Payload) > MaxPayloadSize {
		return ErrPayloadOversized
	}

	// validate gas limit
	if tx.Data.GasLimit < tx.IntrinsicGas() {
		return ErrIntrinsicGas
	}

	if (tx.Data.To.IsEmpty() || tx.Data.To.Type() != common.AddressTypeExternal) && len(tx.Data.Payload) == 0 {
		return ErrPayloadEmpty
	}

	// validate shard of from address
	if shardNeeded && common.IsShardEnabled() {
		if fromShardNum := tx.Data.From.Shard(); fromShardNum != common.LocalShardNumber {
			return fmt.Errorf("invalid from address, shard number is [%v], but coinbase shard number is [%v]", fromShardNum, common.LocalShardNumber)
		}
	}

	// vaildate signature
	if signNeeded {
		if err := tx.verifySignature(); err != nil {
			return err
		}
	}

	return nil
}

// ValidateState validates state dependent fields in tx.
func (tx *Transaction) ValidateState(statedb stateDB, height uint64) error {
	fee := new(big.Int).Mul(tx.Data.GasPrice, new(big.Int).SetUint64(tx.Data.GasLimit))
	cost := new(big.Int).Add(tx.Data.Amount, fee)

	if balance := statedb.GetBalance(tx.Data.From); cost.Cmp(balance) > 0 {
		return fmt.Errorf("balance is not enough, account:%s, balance:%v, amount:%v, fee:%v, cost:%v", tx.Data.From.Hex(), balance, tx.Data.Amount, fee, cost)
	}

	if (height >= common.ThirdForkHeight) {
		if accountNonce := statedb.GetNonce(tx.Data.From); tx.Data.AccountNonce < accountNonce {
			return fmt.Errorf("nonce is too small, account:%s, tx nonce:%d, state db nonce:%d", tx.Data.From.Hex(), tx.Data.AccountNonce, accountNonce)
		}
	} else {
		if accountNonce := statedb.GetNonce(tx.Data.From); tx.Data.AccountNonce < accountNonce {
			return fmt.Errorf("nonce is too small, account:%s, tx nonce:%d, state db nonce:%d", tx.Data.From.Hex(), tx.Data.AccountNonce, accountNonce)
		}
	}

	return nil
}

// IntrinsicGas computes the 'intrinsic gas' for a tx.
func (tx *Transaction) IntrinsicGas() uint64 {
	gas := ethIntrinsicGas(tx.Data.Payload)

	if tx.IsCrossShardTx() {
		return gas * 2
	}

	return gas
}


// IsCrossShardTx indicates whether the tx is to another shard.
func (tx *Transaction) IsCrossShardTx() bool {
	return !tx.Data.From.IsEmpty() && !tx.Data.To.IsEmpty() && !tx.Data.To.IsReserved() && tx.Data.From.Shard() != tx.Data.To.Shard()
}

func (tx *Transaction) verifySignature() error {
	if len(tx.Signature.Sig) == 0 {
		return ErrSigMissing
	}

	if txHash := tx.CalculateHash(); !txHash.Equal(tx.Hash) {
		return ErrHashMismatch
	}

	// cache key is made up of tx hash and signature
	key := string(append(tx.Hash.Bytes(), tx.Signature.Sig...))

	if v, ok := sigCache.Get(key); ok {
		if v == nil {
			return nil
		}

		return v.(error)
	}

	var err error
	if !tx.Signature.Verify(tx.Data.From, tx.Hash.Bytes()) {
		err = ErrSigInvalid
	}

	// verify signature is time consuming, so cache the result.
	sigCache.Add(key, err)

	return err
}


// ethIntrinsicGas computes the 'intrinsic gas' for a message with the given data.
func ethIntrinsicGas(data []byte) uint64 {
	// Set the starting gas for the raw transaction
	gas := params.TxGas

	if len(data) == 0 {
		return gas
	}

	// Bump the required gas by the amount of transactional data
	// Zero and non-zero bytes are priced differently
	var nz uint64
	for _, byt := range data {
		if byt != 0 {
			nz++
		}
	}

	// will not overflow, since maximum tx payload size is 32K.
	gas += nz * params.TxDataNonZeroGas
	z := uint64(len(data)) - nz
	gas += z * params.TxDataZeroGas

	return gas
}

// Validate validates all fields in tx.
func (tx *Transaction) Validate(statedb stateDB, height uint64) error {
	if err := tx.ValidateWithoutState(true, true); err != nil {
		return err
	}

	return tx.ValidateState(statedb, height)
}