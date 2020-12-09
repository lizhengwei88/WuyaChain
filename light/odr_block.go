package light

import (
	"WuyaChain/common"
	"WuyaChain/core/types"
)

type odrBlock struct {
	OdrItem
	Hash  common.Hash  // Retrieved block hash
	Block *types.Block `rlp:"nil"` // Retrieved block
}