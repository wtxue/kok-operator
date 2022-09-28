package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2/klogr"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Execute executes the current command using RootCmd.
func Execute() {
	gOpt := defaultGlobalManagerOption()

	rootCmd := &cobra.Command{
		Use:   "ctl",
		Short: "ctl is a command line interface ",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			logf.SetLogger(klogr.New())
		},
	}

	gOpt.AddFlags(rootCmd.PersistentFlags())

	rootCmd.AddCommand(NewToolsCmd(gOpt))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
