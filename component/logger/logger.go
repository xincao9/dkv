package logger

import (
	"dkv/component/constant"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"log"
	"path/filepath"
)

const (
	maxSize    = 500
	maxBackups = 3
	maxAge     = 7
	compress   = true
	file       = "dkv.log"
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
	fn := filepath.Join(constant.LoggerDir, file)
	L.Out = &lumberjack.Logger{
		Filename:   fn,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	L.SetLevel(level)
	L.Formatter = &logrus.JSONFormatter{}
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(L.WriterLevel(logrus.InfoLevel))
}
