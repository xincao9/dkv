package logger

import (
	"dkv/store/config"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"log"
	"path/filepath"
)

var (
	D *logrus.Logger
)

func init() {
	// 日志设置
	D = logrus.New()
	level, err := logrus.ParseLevel(config.D.GetString("logger.level"))
	if err != nil {
		log.Fatalf("Fatal error config file logger.level: %v\n", err)
	}
	fn := filepath.Join(config.D.GetString("data.dir"), "server.log")
	D.Out = &lumberjack.Logger{
		Filename:   fn,
		MaxSize:    500,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}
	D.SetLevel(level)
	D.Formatter = &logrus.JSONFormatter{}
	log.SetFlags(log.LstdFlags | log.Llongfile)
	log.SetOutput(D.WriterLevel(logrus.InfoLevel))
}
