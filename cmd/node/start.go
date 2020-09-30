package node

import (
	"fmt"
	"github.com/spf13/cobra"
)

var wuyaNodeconfigFile string

var startCmd=&cobra.Command{
	Use: "start",
	Short: " command for starting a Wuya.node",
	Long: `usage example:
		cmd.exe start -c config\node.json
		start a wuyanode.`,
	Run: func(cmd *cobra.Command, args []string) {
		 //do stuff here
    nodeConfig,err:= LoadConfigFromFile(wuyaNodeconfigFile)
	if err !=nil{
		fmt.Printf("failed to reading the config file:%s\n", err)
		return
	}
	fmt.Println("nodeConfig:", nodeConfig)
	},
}


func init()  {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&wuyaNodeconfigFile,"config","c","","wuya node config file (required)")
	startCmd.MarkFlagRequired("config")
}