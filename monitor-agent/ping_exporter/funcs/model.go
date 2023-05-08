package funcs

import "log"

type TelnetObj struct {
	Ip      string
	Port    int
	Success bool
}

type HttpCheckObj struct {
	Method     string
	Url        string
	StatusCode int
}

func DebugLog(msg string, v ...interface{}) {
	if Config().Debug {
		msg = msg + " \n"
		log.Printf(msg, v...)
	}
}

type PingResultObj struct {
	Ip          string
	UpDown      int
	UseTime     float64
	LossPercent float64
}
