/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package light

import (
	"WuyaChain/common"
	"WuyaChain/core/store"
	"WuyaChain/core/types"
)

// ODR object to send tx.
type odrAddTx struct {
	OdrItem
	Tx types.Transaction
}

func (odr *odrAddTx) code() uint16 {
	return addTxRequestCode
}


func (odr *odrAddTx) validate(request odrRequest, bcStore store.BlockchainStore) error {
	return nil
}

// ODR object to get transaction by hash.
type odrTxByHashRequest struct {
	OdrItem
	TxHash common.Hash
}

func (req *odrTxByHashRequest) code() uint16 {
	return txByHashRequestCode
}
