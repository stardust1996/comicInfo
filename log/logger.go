package log

import (
	"fmt"
	"log"
	"os"
)

var Logger *log.Logger

/**
 * 2024/2/5
 * add by stardust
**/

func init() {
	logFile, err := os.OpenFile("log.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("创建日志文件失败", err)
		return
	}
	Logger = log.New(logFile, "", log.Llongfile|log.Lmicroseconds|log.Ldate)
}
