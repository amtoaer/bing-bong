package utils

import log "github.com/sirupsen/logrus"

// 检测一系列错误，如发现错误则退出程序
func Fatalf(msg string, errs ...error) {
	for _, err := range errs {
		if err != nil {
			log.Fatalf(msg, err)
		}
	}
}

func Errorf(msg string, errs ...error) bool {
	var flag bool
	for _, err := range errs {
		if err != nil {
			log.Errorf(msg, err)
			flag = true
		}
	}
	return flag
}
