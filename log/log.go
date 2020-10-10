package log

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"go-serverRPC/log/comm"
	"os"
	"path/filepath"
	"time"
)

const logExtension =".log"

type WuyaLog struct {
	log *logrus.Logger
}
var (
	LogFolder=filepath.Join()
	logMap map[string]*WuyaLog
)

func (l *WuyaLog) Info(format string,args ...interface{})  {
	l.log.Infof(format,args)
}

func (l *WuyaLog) Error(format string,args ...interface{})  {
	l.log.Error(format,args)
}

func GetLogger(module string) *WuyaLog {
    if logMap==nil{
    	logMap=make(map[string]*WuyaLog)
	}
    curLog,ok:=logMap[module]
	if ok{
		return  curLog
	}
   logrus.SetFormatter(&logrus.TextFormatter{})
    log:=logrus.New()

    logDir:=filepath.Join(LogFolder,comm.LogConfiguration.DataDir)
    err:=os.MkdirAll(logDir,os.ModePerm)
    if err!=nil{
    	panic(fmt.Sprintf("failed to create log dir:%s",err.Error()))
	}
    logFileName:=fmt.Sprintf("%s%s","%Y%m%d",logExtension)

    writer,err:=rotatelogs.New(
    	filepath.Join(logDir,logFileName),
    	rotatelogs.WithClock(rotatelogs.Local),
    	rotatelogs.WithMaxAge(24*7*time.Hour),
    	rotatelogs.WithRotationTime(24*time.Hour),
    	)
    if err!=nil{
    	panic(fmt.Sprintf("failed to create log file:%s",err))
	}
	log.Out=writer
	log.SetLevel(logrus.DebugLevel)
	curLog=&WuyaLog{
		log: log,
	}
	logMap[module]=curLog
    return curLog
}