package node

import (
	"WuyaChain/common"
	"WuyaChain/consensus"
	"WuyaChain/consensus/factory"
	"WuyaChain/lightclients"
	"WuyaChain/log"
	"WuyaChain/node"
	"WuyaChain/wuya"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"sync"
)

var (
	wuyaNodeconfigFile string
	miner               string
	metricsEnableFlag   bool
	accountsConfig      string
	threads             int
	startHeight         int
	percentage          int

	// default is full node
	lightNode bool

	//pprofPort http server port
	pprofPort uint64

	// profileSize is used to limit when need to collect profiles, set 6GB
	profileSize = uint64(1024 * 1024 * 1024 * 6)

	maxConns       = int(0)
	maxActiveConns = int(0)
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: " command for starting a Wuya.node",
	Long: `usage example:
		cmd.exe start -c config\node.json
		start a wuyanode.`,
	Run: func(cmd *cobra.Command, args []string) {
		//do stuff here
		var wg sync.WaitGroup
		nodeConfig, err := LoadConfigFromFile(wuyaNodeconfigFile)
		if err != nil {
			fmt.Printf("failed to reading the config file:%s\n", err)
			return
		}
		// Create log
		wlog := log.GetLogger("wuya")
		serviceContext := wuya.ServiceContext{
			DataDir: nodeConfig.BasicConfig.DataDir,
		}
		ctx := context.WithValue(context.Background(), "ServiceContext", serviceContext)

		wuyaNode, err := node.NewPToP(nodeConfig)
		if err != nil {
			fmt.Printf("failed to reading the config file:%s\n", err)
			return
		}

		var engine consensus.Engine
		if nodeConfig.BasicConfig.MinerAlgorithm == common.BFTEngine {
			//engine, err = factory.GetBFTEngine(nodeConfig.WuyaConfig.CoinbasePrivateKey, nodeConfig.BasicConfig.DataDir)
		} else {
			engine, err = factory.GetConsensusEngine(nodeConfig.BasicConfig.MinerAlgorithm, nodeConfig.BasicConfig.DataSetDir, percentage)
		}

		// light client manager
		_, err = lightclients.NewLightClientManager(wuyaNode.GetShardNumber(), ctx, nodeConfig, engine)
		if err != nil {
			fmt.Printf("create light client manager failed. %s", err)
			return
		}

		// fullnode mode
		wuyaService, err := wuya.NewWuyaService(ctx, nodeConfig, wlog,engine, startHeight)
		if err != nil {
			return
		}

		err = wuyaNode.Start()
		if err != nil {
			fmt.Print("got error when start node:%s\n", err)
			return
		}

		wuyaService.Miner().Start()
		wg.Add(1)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	startCmd.Flags().StringVarP(&wuyaNodeconfigFile, "config", "c", "", "seele node config file (required)")

	startCmd.Flags().StringVarP(&miner, "miner", "m", "start", "miner start or not, [start, stop]")
	startCmd.Flags().BoolVarP(&metricsEnableFlag, "metrics", "t", false, "start metrics")
	startCmd.Flags().StringVarP(&accountsConfig, "accounts", "", "", "init accounts info")
	startCmd.Flags().IntVarP(&threads, "threads", "", 1, "miner thread value")
	startCmd.Flags().IntVarP(&percentage, "percentage", "p", 10, "miner target confidence range (integer, 1-10), higher: more calculation time; lower: less chance to hit target")
	startCmd.Flags().BoolVarP(&lightNode, "light", "l", false, "whether start with light mode")
	startCmd.Flags().Uint64VarP(&pprofPort, "port", "", 0, "which port pprof http server listen to")
	startCmd.Flags().IntVarP(&startHeight, "startheight", "", -1, "the block height to start from")
	startCmd.Flags().IntVarP(&maxConns, "maxConns", "", 0, "node max connections")
	startCmd.Flags().IntVarP(&maxActiveConns, "maxActiveConns", "", 0, "node max active connections")
}
