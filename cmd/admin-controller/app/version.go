package app

import (
	"github.com/spf13/cobra"

	"fmt"
	"os"

	"github.com/wtxue/kube-on-kube-operator/pkg/version"
)

// NewCmdVersion returns a cobra command for fetching versions
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  "Print the version information for the current context",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.GetVersion()
			fmt.Fprintf(os.Stdout, "version: %v\n", v.String())
		},
	}

	return cmd
}
