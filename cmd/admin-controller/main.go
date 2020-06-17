package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/wtxue/kube-on-kube-operator/cmd/admin-controller/app"
	_ "github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/cluster"
	_ "github.com/wtxue/kube-on-kube-operator/pkg/provider/baremetal/machine"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	rootCmd := app.GetRootCmd(os.Args[1:])

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}
}
