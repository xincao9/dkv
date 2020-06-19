package logger

import (
    "dkv/component/constant"
    "github.com/natefinch/lumberjack"
    "github.com/sirupsen/logrus"
    "log"
    "path/filepath"
)

var (
	L *logrus.Logger
)

func init() {
	// 日志设置
	L = logrus.New()
	level, err := logrus.ParseLevel(constant.LoggerLevel)
	if err != nil {
		log.Fatalf("Fatal error logger : %v\n", err)
	}
	fn := filepath.Join(constant.Dir, "server.log")
	L.Out = &lumberjack.Logger{
		Filename:   fn,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	L.SetLevel(level)
	L.Formatter = &logrus.JSONFormatter{}
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(L.WriterLevel(logrus.InfoLevel))
}
