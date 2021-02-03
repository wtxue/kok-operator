package gmanager

import (
	"github.com/wtxue/kok-operator/pkg/clustermanager"
	"github.com/wtxue/kok-operator/pkg/provider"
	"github.com/wtxue/kok-operator/pkg/provider/config"
)

type GManager struct {
	*provider.ProviderManager
	*clustermanager.ClusterManager
	*config.Config
}
