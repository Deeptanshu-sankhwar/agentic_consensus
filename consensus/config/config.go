package config

import (
	"os"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
)

func DefaultConfig(rootDir string) *cfg.Config {
	config := cfg.DefaultConfig()
	config.BaseConfig.RootDir = rootDir
	config.BaseConfig.ProxyApp = "tcp://127.0.0.1:26658"
	config.P2P.ListenAddress = "tcp://0.0.0.0:26656"
	config.P2P.AllowDuplicateIP = true
	config.Consensus.TimeoutCommit = 1000
	config.Consensus.SkipTimeoutCommit = false
	config.RPC.ListenAddress = "tcp://0.0.0.0:26657"
	return config
}

func InitFilesWithConfig(config *cfg.Config) error {
	cfg.EnsureRoot(config.RootDir)
	privValKeyFile := config.PrivValidatorKeyFile()
	privValStateFile := config.PrivValidatorStateFile()

	if !tExists(privValKeyFile) {
		pv := privval.GenFilePV(privValKeyFile, privValStateFile)
		pv.Save()
	}

	nodeKeyFile := config.NodeKeyFile()
	_, err := p2p.LoadOrGenNodeKey(nodeKeyFile)
	if err != nil {
		return err
	}

	return nil
}

func tExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
