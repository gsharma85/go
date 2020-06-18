package utils

import (
	"log"
	"os"
)

func CreateFileLogger(fileName string) log.*Logger {
	if logger == nil {
		logFile, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
		   panic(err)
		}		
		logger := log.New(logFile, "", log.LstdFlags)
	}
	
	return logger;
}
