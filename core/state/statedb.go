package state

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
	"WuyaChain/database"
	"WuyaChain/trie"
	"bytes"
	"math/big"
)


var (
	// TrieDbPrefix is the key prefix of trie database in statedb.
	TrieDbPrefix  = []byte("S")
	stateBalance0 = big.NewInt(0)
)


// Trie is used for statedb to store key-value pairs.
// For full node, it's MPT based on levelDB.
// For light node, it's a ODR trie with limited functions.
type Trie interface {
	Hash() common.Hash
	Commit(batch database.Batch) common.Hash
	Get(key []byte) ([]byte, bool, error)
	Put(key, value []byte) error
	DeletePrefix(prefix []byte) (bool, error)
	GetProof(key []byte) (map[string][]byte, error)
}

// Statedb is used to store accounts into the MPT tree
type Statedb struct {
	trie         Trie
	stateObjects map[common.Address]*stateObject

	dbErr  error  // dbErr is used for record the database error.
	refund uint64 // The refund counter, also used by state transitioning.

	// Receipt logs for current processed tx.
	curTxIndex uint
	curLogs    []*types.Log

	// State modifications for current processed tx.
	curJournal *journal
}

// NewStatedb constructs and returns a statedb instance
func NewStatedb(root common.Hash, db database.Database) (*Statedb, error) {
	trie, err := trie.NewTrie(root, TrieDbPrefix, db)
	if err != nil {
		return nil, err
	}

	return NewStatedbWithTrie(trie), nil
}


// NewEmptyStatedb creates an empty statedb instance.
func NewEmptyStatedb(db database.Database) *Statedb {
	trie := trie.NewEmptyTrie(TrieDbPrefix, db)
	return NewStatedbWithTrie(trie)
}


// NewStatedbWithTrie creates a statedb instance with specified trie.
func NewStatedbWithTrie(trie Trie) *Statedb {
	return &Statedb{
		trie:         trie,
		stateObjects: make(map[common.Address]*stateObject),
		curJournal:   newJournal(),
	}
}

// CreateAccount creates a new account in statedb.
func (s *Statedb) CreateAccount(address common.Address) {
	if object := s.getStateObject(address); object == nil {
		object = newStateObject(address)
		s.curJournal.append(createObjectChange{&address})
		s.stateObjects[address] = object
	}
}

func (s *Statedb) getStateObject(addr common.Address) *stateObject {
	// get from cache
	if object, ok := s.stateObjects[addr]; ok {
		if !object.deleted {
			return object
		}

		// object has already been deleted from trie.
		return nil
	}

	// load from trie
	object := newStateObject(addr)
	ok, err := object.loadAccount(s.trie)
	if err != nil {
		s.setError(err)
	}

	if err != nil || !ok {
		return nil
	}

	// add in cache
	s.stateObjects[addr] = object

	return object
}

// setError only records the first error.
func (s *Statedb) setError(err error) {
	if s.dbErr == nil {
		s.dbErr = err
	}
}

// SetBalance sets the balance of the specified account if exists.
func (s *Statedb) SetBalance(addr common.Address, balance *big.Int) {
	if object := s.getStateObject(addr); object != nil {
		s.curJournal.append(balanceChange{&addr, object.getAmount()})
		object.setAmount(balance)
	}
}


// AddBalance adds the specified amount to the balance for the specified account if exists.
func (s *Statedb) AddBalance(addr common.Address, amount *big.Int) {
	if object := s.getStateObject(addr); object != nil {
		s.curJournal.append(balanceChange{&addr, object.getAmount()})
		object.addAmount(amount)
	}
}

// GetNonce gets the nonce of the specified account if exists.
// Otherwise, return 0.
func (s *Statedb) GetNonce(addr common.Address) uint64 {
	if object := s.getStateObject(addr); object != nil {
		return object.getNonce()
	}

	return 0
}

// SetNonce sets the nonce of the specified account if exists.
func (s *Statedb) SetNonce(addr common.Address, nonce uint64) {
	if object := s.getStateObject(addr); object != nil {
		s.curJournal.append(nonceChange{&addr, object.getNonce()})
		object.setNonce(nonce)
	}
}

// Hash flush the dirty data into trie and calculates the intermediate root hash.
func (s *Statedb) Hash() (common.Hash, error) {
	if s.dbErr != nil {
		return common.EmptyHash, s.dbErr
	}

	for addr := range s.curJournal.dirties {
		if object, found := s.stateObjects[addr]; found {
			if err := object.flush(s.trie); err != nil {
				return common.EmptyHash, err
			}
		}
	}

	s.clearJournalAndRefund()

	return s.trie.Hash(), nil
}

// Commit persists the trie to the specified batch.
func (s *Statedb) Commit(batch database.Batch) (common.Hash, error) {
	if batch == nil {
		panic("batch is nil")
	}

	if s.dbErr != nil {
		return common.EmptyHash, s.dbErr
	}

	for _, object := range s.stateObjects {
		if err := object.flush(s.trie); err != nil {
			return common.EmptyHash, err
		}
	}
	return s.trie.Commit(batch), nil
}

// SubBalance subtracts the specified amount from the balance for the specified account if exists.
func (s *Statedb) SubBalance(addr common.Address, amount *big.Int) {
	if object := s.getStateObject(addr); object != nil {
		s.curJournal.append(balanceChange{&addr, object.getAmount()})
		object.subAmount(amount)
	}
}


func (s *Statedb) clearJournalAndRefund() {
	s.refund = 0
	s.curJournal.entries = s.curJournal.entries[:0]
	s.curJournal.dirties = make(map[common.Address]uint)
}

// Prepare resets the logs and journal to process a new tx and return the statedb snapshot.
func (s *Statedb) Prepare(txIndex int) int {
	s.curTxIndex = uint(txIndex)
	s.curLogs = nil

	s.clearJournalAndRefund()
	return s.Snapshot()
}

// Snapshot returns an identifier for the current revision of the statedb.
func (s *Statedb) Snapshot() int {
	return s.curJournal.snapshot()
}

// GetData returns the account data of the specified key if exists.
// Otherwise, return nil.
func (s *Statedb) GetData(addr common.Address, key common.Hash) []byte {
	return s.getData(addr, key, false)
}

func (s *Statedb) getData(addr common.Address, key common.Hash, committed bool) []byte {
	object := s.getStateObject(addr)
	if object == nil {
		return nil
	}

	data, err := object.getState(s.trie, key, committed)
	if err != nil {
		s.setError(err)
	}

	return data
}

// SetData sets the key value pair for the specified account if exists.
func (s *Statedb) SetData(addr common.Address, key common.Hash, value []byte) {
	object := s.getStateObject(addr)
	if object == nil {
		return
	}

	prevValue, err := object.getState(s.trie, key, false)
	if err != nil {
		s.setError(err)
		return
	}

	// value not changed.
	if bytes.Equal(prevValue, value) {
		return
	}

	s.curJournal.append(storageChange{&addr, key, prevValue})
	object.setState(key, value)
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (s *Statedb) RevertToSnapshot(revid int) {
	s.curJournal.revert(s, revid)
}

// GetRefund returns the current value of the refund counter.
func (s *Statedb) GetRefund() uint64 {
	return s.refund
}

// Exist indicates whether the given account exists in statedb.
// Note that it should also return true for suicided accounts.
func (s *Statedb) Exist(address common.Address) bool {
	return s.getStateObject(address) != nil
}

// GetBalance returns the balance of the specified account if exists.
// Otherwise, returns zero.
func (s *Statedb) GetBalance(addr common.Address) *big.Int {
	if object := s.getStateObject(addr); object != nil {
		return object.getAmount()
	}

	return stateBalance0
}

// GetCurrentLogs returns the current transaction logs.
func (s *Statedb) GetCurrentLogs() []*types.Log {
	return s.curLogs
}
