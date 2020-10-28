package node

import (
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
	startHeight        int
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

		wuyaService, err := wuya.NewWuyaService(ctx, nodeConfig, wlog,startHeight)
		if err != nil {
			fmt.Println("wuyaService:", err.Error())
			return
		}

		fmt.Println("wuyaService:", wuyaService)
		err = wuyaNode.Start()
		if err != nil {
			fmt.Print("got error when start node:%s\n", err)
			return
		}

	//	wuyaService.Miner().Start()
		wg.Add(1)
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&wuyaNodeconfigFile, "config", "c", "", "wuya node config file (required)")
	startCmd.Flags().IntVarP(&startHeight, "startheight", "", -1, "the block height to start from")
	startCmd.MarkFlagRequired("config")
}
