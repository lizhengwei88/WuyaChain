package node

import (
	"WuyaChain/cmd/util"
	"WuyaChain/node"
	"encoding/json"
	"io/ioutil"
)

func LoadConfigFromFile(configFile string) (*node.Config, error) {
	cmdConfig,err:= GetConfigFromFile(configFile)
	if err!=nil{
		return nil,err
	}
	config:=CopyConfig(cmdConfig)
	return config,err
}

func GetConfigFromFile(filepath string)  (*util.Config,error)  {
	var config util.Config
	 buff,err:= ioutil.ReadFile(filepath)
	 if err!=nil{
		 return &config,err
	 }
	 err=json.Unmarshal(buff,&config)
	return &config,err
}

func CopyConfig(cmdConfig *util.Config) *node.Config  {
	config:=&node.Config{
		BasicConfig: cmdConfig.BasicConfig,
	}
	return config
}