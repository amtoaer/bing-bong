package utils

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logger = logrus.New()

func InitLog() {
	logFile, err := os.OpenFile(viper.GetString("log"), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logger.Fatalf("打开日志文件出现错误：%v", err)
	}
	logger.SetOutput(logFile)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args)
}

func Info(args ...interface{}) {
	logger.Info(args)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args)
}
