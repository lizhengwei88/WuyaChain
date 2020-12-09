package node

import (
	"WuyaChain/cmd/util"
	"WuyaChain/common"
	"WuyaChain/node"
	"encoding/json"
	"WuyaChain/crypto"
	"io/ioutil"
)

func LoadConfigFromFile(configFile string) (*node.Config, error) {
	cmdConfig, err := GetConfigFromFile(configFile)
	if err != nil {
		return nil, err
	}
	config := CopyConfig(cmdConfig)

	if len(config.BasicConfig.Coinbase) > 0 {
		config.WuyaConfig.Coinbase = common.HexMustToAddres(config.BasicConfig.Coinbase)
	}

	if cmdConfig.P2PConfig.PrivateKey == nil {
		key, err := crypto.LoadECDSAFromString(cmdConfig.P2PConfig.SubPrivateKey)
		if err != nil {
			return config, err
		}
		config.P2PConfig.PrivateKey = key
	}

    config.WuyaConfig.GenesisConfig=cmdConfig.GenesisConfig
	return config, err
}

func GetConfigFromFile(filepath string) (*util.Config, error) {
	var config util.Config
	buff, err := ioutil.ReadFile(filepath)
	if err != nil {
		return &config, err
	}
	err = json.Unmarshal(buff, &config)
	return &config, err
}

func CopyConfig(cmdConfig *util.Config) *node.Config {
	config := &node.Config{
		BasicConfig: cmdConfig.BasicConfig,
		P2PConfig:   cmdConfig.P2PConfig,
		WuyaConfig: node.WuyaConfig{},
	}
	return config
}
