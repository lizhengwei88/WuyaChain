package common

import (
	"math/big"
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	WindowsPipeDir = `\\.\pipe\`

	defaultPipeFile = `\seele.ipc`
	// EthashAlgorithm miner algorithm ethash
	EthashAlgorithm = "ethash"
	// Sha256Algorithm miner algorithm sha256
	Sha256Algorithm = "sha256"
	// spow miner algorithm
	SpowAlgorithm = "spow"

	// ShardCount represents the total number of shards.
	ShardCount = 4
	// ForkHeight after this height we change the content of block: hardFork
	ForkHeight = 130000
	// ForkHeight after this height we change the content of block: hardFork
	SecondForkHeight = 145000
	// BFT mineralgorithm
	BFTEngine = "bft"
	// BFT data folder
	BFTDataFolder = "bftdata"
	// ForkHeight after this height we change the validation of tx: hardFork
	ThirdForkHeight = 735000
	SmartContractNonceForkHeight = 1100000
	HeightRoof  = uint64(707996)
	HeightFloor = uint64(707989)
)
var (
    tempFolder string
	// defaultDataFolder used to store persistent data info, such as the database and keystore
	defaultDataFolder string
	// defaultIPCPath used to store the ipc file
	defaultIPCPath string
)

var (
	Big0   = big.NewInt(0)
)

func init()  {
	usr,err:=user.Current()
    if err!=nil{
     panic(err)
	}
   tempFolder=filepath.Join(usr.HomeDir,"wuyaTemp")
   defaultDataFolder=filepath.Join(usr.HomeDir,".wuya")
   if runtime.GOOS=="windows"{
   	defaultIPCPath=WindowsPipeDir+defaultPipeFile
   }else {
   	defaultIPCPath=filepath.Join(defaultDataFolder,defaultPipeFile)
   }
}

// GetTempFolder uses a getter to implement readonly
func GetTempFolder() string {
	return tempFolder
}
