package funcs

import "log"

type TelnetObj struct {
	Ip  string
	Port  int
	Success  bool
}

func DebugLog(msg string, v ...interface{}){
	if Config().Debug {
		msg = msg + " \n"
		log.Printf(msg, v...)
	}
}
