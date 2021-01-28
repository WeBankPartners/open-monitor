package log

import (
	"encoding/json"
	m "github.com/WeBankPartners/open-monitor/monitor-server/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"strings"
	"time"
)

var (
	Logger *zap.Logger
	AccessLogger *zap.Logger
	levelStringList = []string{"debug","info","warn","error"}
	LogLevel string
)

func InitLogger()  {
	Logger = InitArchiveZapLogger(m.Config().Http.Log.File)
	AccessLogger = InitArchiveZapLogger(m.Config().Http.Log.AccessFile)
}

func InitArchiveZapLogger(filePath string) *zap.Logger {
	LogLevel = strings.ToLower(m.Config().Http.Log.Level)
	var level int
	for i,v := range levelStringList {
		if v == LogLevel {
			level = i-1
			break
		}
	}
	zapLevel := zap.NewAtomicLevel()
	zapLevel.SetLevel(zapcore.Level(level))
	hook := lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    m.Config().Http.Log.ArchiveMaxSize,
		MaxBackups: m.Config().Http.Log.ArchiveMaxBackup,
		MaxAge:     m.Config().Http.Log.ArchiveMaxDay,
		Compress:   m.Config().Http.Log.Compress,
	}
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(&hook)), zapLevel)
	logger := zap.New(core, zap.AddCaller(), zap.Development())
	logger.Info("success init zap log !!")
	return logger
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func String(k,v string) zap.Field {
	return zap.String(k, v)
}

func Int(k string,v int) zap.Field {
	return zap.Int(k, v)
}

func Int64(k string,v int64) zap.Field {
	return zap.Int64(k, v)
}

func Float64(k string,v float64) zap.Field {
	return zap.Float64(k, v)
}

func JsonObj(k string,v interface{}) zap.Field {
	b,err := json.Marshal(v)
	if err == nil {
		return zap.String(k, string(b))
	}else{
		return zap.Error(err)
	}
}

func StringList(k string,v []string) zap.Field {
	return zap.Strings(k, v)
}