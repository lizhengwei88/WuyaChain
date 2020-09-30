package node

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "node",
	Short: "node command for starting a node",
	Long:  `use "node help [<command>]" for detailed usage`,

}

 func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
