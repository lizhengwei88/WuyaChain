package node

import (
	"WuyaChain/log"
	"WuyaChain/node"
	"WuyaChain/wuya"
	"context"
	"fmt"
	"github.com/spf13/cobra"
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
		nodeConfig, err := LoadConfigFromFile(wuyaNodeconfigFile)
		if err != nil {
			fmt.Printf("failed to reading the config file:%s\n", err)
			return
		}
		// Create log
		wlog := log.GetLogger("wuya")
		fmt.Println("log is success 001")
		serviceContext := wuya.ServiceContext{
			DataDir: nodeConfig.BasicConfig.DataDir,
		}
		ctx := context.WithValue(context.Background(), "ServiceContext", serviceContext)
		fmt.Println("ctx is success 001")
		wuyaNode, err := node.NewPToP(nodeConfig)
		if err != nil {
			fmt.Printf("failed to reading the config file:%s\n", err)
			return
		}
		fmt.Println("NewPToP is success 001")
		wuyaService, err := wuya.NewWuyaService(ctx, nodeConfig, wlog)
		if err != nil {
			fmt.Println("wuyaService:",err.Error())
			return
		}

		fmt.Println("wuyaNode:",wuyaNode)
		fmt.Println("wuyaService:",wuyaService)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&wuyaNodeconfigFile, "config", "c", "", "wuya node config file (required)")
	startCmd.MarkFlagRequired("config")
}
