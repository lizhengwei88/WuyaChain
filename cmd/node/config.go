package node

import (
	"WuyaChain/cmd/util"
	"encoding/json"
	"io/ioutil"
)

func LoadConfigFromFile(configFile string) (*util.Config,error)  {
	cmdConfig,err:= GetConfigFromFile(configFile)
	if err!=nil{
		return nil,err
	}
	return cmdConfig,err
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