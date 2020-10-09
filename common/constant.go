package common

import (
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	WindowsPipeDir = `\\.\pipe\`

	defaultPipeFile = `\seele.ipc`
)
var (
    tempFolder string
	// defaultDataFolder used to store persistent data info, such as the database and keystore
	defaultDataFolder string
	// defaultIPCPath used to store the ipc file
	defaultIPCPath string
)

func init()  {
	usr,err:=user.Current()
    if err!=nil{
     panic(err)
	}
   tempFolder=filepath.Join(usr.HomeDir,"wuyaTemp")
   defaultDataFolder=filepath.Join(usr.HomeDir,".wuya")
   if runtime.GOOS=="windows"{
   	defaultIPCPath=WindowsPipeDir+defaultPipeFile
   }else {
   	defaultIPCPath=filepath.Join(defaultDataFolder,defaultPipeFile)
   }
}

// GetTempFolder uses a getter to implement readonly
func GetTempFolder() string {
	return tempFolder
}
