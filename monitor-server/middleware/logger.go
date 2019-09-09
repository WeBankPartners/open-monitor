package middleware

import (
	"go.uber.org/zap"
	"fmt"
	"go.uber.org/zap/zapcore"
	m "github.com/WeBankPartners/wecube-plugins-prometheus/monitor-server/models"
	"strings"
	"log"
)

var (
	LogHttp *zap.Logger
	//LogCron *zap.Logger
	HttpLogEnable  bool
)

func InitHttpLog()  {
	HttpLogEnable = m.Config().Http.Log.Enable
	//enable := true
	if !HttpLogEnable {
		return
	}
	var zCfg zap.Config
	// debug:-1 info:0 warn:1 error:2 DPanic:3 Panic:4 Fatal:5
	zCfg.Level = zap.NewAtomicLevelAt(-1)
	zCfg.Encoding = "json"
	zCfg.OutputPaths = []string{}
	zCfg.ErrorOutputPaths = []string{}
	if m.Config().Http.Log.Stdout {
		zCfg.OutputPaths = append(zCfg.OutputPaths, "stdout")
		zCfg.ErrorOutputPaths = append(zCfg.ErrorOutputPaths, "stderr")
	}
	cfgFile := m.Config().Http.Log.File
	//cfgFile := "/app/data/log/monitor.log"
	if cfgFile!="" {
		zCfg.OutputPaths = append(zCfg.OutputPaths, cfgFile)
		cfgErrorFile := cfgFile + `_err`
		if strings.Contains(cfgFile, ".log") {
			cfgErrorFile = strings.TrimRight(cfgFile, ".log") + `_err.log`
		}
		zCfg.ErrorOutputPaths = append(zCfg.ErrorOutputPaths, cfgErrorFile)
	}
	//initialFieldsMap := make(map[string]interface{})
	//initialFieldsMap["const_key"] = "const_val"
	//zCfg.InitialFields = initialFieldsMap
	var encoderMap zapcore.EncoderConfig
	encoderMap.MessageKey = "msg"
	encoderMap.LevelKey = "level"
	zCfg.EncoderConfig = encoderMap
	zCfg.EncoderConfig = zap.NewProductionEncoderConfig()
	zCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	LogHttp, err = zCfg.Build()
	if err!=nil {
		fmt.Println("error init logger!! logger build fail ")
		return
	}
	LogHttp.Info("success init zap log !!")
	defer LogHttp.Sync()
}

func LogError(s string, e error) {
	if HttpLogEnable {
		LogHttp.Error(s, zap.Error(e))
	}else{
		log.Printf("%s  :  %v \n", s, e)
	}
}

func LogInfo(s string) {
	if HttpLogEnable {
		LogHttp.Info(s)
	}else{
		log.Println(s)
	}
}
