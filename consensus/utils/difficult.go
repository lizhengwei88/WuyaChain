/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package utils

import (
	"WuyaChain/common"
	"WuyaChain/consensus"
	"WuyaChain/core/types"

	"math/big"


)

// getDifficult adjust difficult by parent info
func GetDifficult(time uint64, parentHeader *types.BlockHeader) *big.Int {
	// algorithm:
	// diff = parentDiff + parentDiff / 2048 * max (1 - (blockTime - parentTime) / 10, -99)
	// target block time is 10 seconds
	parentDifficult := parentHeader.Difficulty
	parentTime := parentHeader.CreateTimestamp.Uint64()
	if parentHeader.Height == 0 {
		return parentDifficult
	}

	big1 := big.NewInt(1)
	big99 := big.NewInt(-99)
	big1024 := big.NewInt(1024)
	big2048 := big.NewInt(2048)

	interval := (time - parentTime) / 10
	var x *big.Int
	x = big.NewInt(int64(interval))
	x.Sub(big1, x)
	if x.Cmp(big99) < 0 {
		x = big99
	}

	var y = new(big.Int).Set(parentDifficult)
	if parentHeader.Height < common.SecondForkHeight {
		y.Div(parentDifficult, big2048)
	} else {
		y.Div(parentDifficult, big1024)
	}

	var result = big.NewInt(0)
	result.Mul(x, y)
	result.Add(parentDifficult, result)

	// fork control for shard 1
	bigUpperLimit := big.NewInt(10000000)
	if parentHeader.Creator.Shard() == uint(1) && parentHeader.Height == common.ForkHeight && result.Cmp(bigUpperLimit) > 0 {
		result = bigUpperLimit
	}

	return result
}

func VerifyDifficulty(parent *types.BlockHeader, header *types.BlockHeader) error {
	difficult := GetDifficult(header.CreateTimestamp.Uint64(), parent)
	if difficult.Cmp(header.Difficulty) != 0 {
		return consensus.ErrBlockDifficultInvalid
	}

	return nil
}
