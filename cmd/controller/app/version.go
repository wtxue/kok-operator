package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wtxue/kok-operator/pkg/version"
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
