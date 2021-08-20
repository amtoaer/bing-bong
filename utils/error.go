package utils

import "log"

// 检测一系列错误，如发现错误则退出程序
func CheckError(msg string, errs ...error) {
	for _, err := range errs {
		if err != nil {
			log.Fatalf(msg, err)
		}
	}
}
