package cache

import (
	"dkv/logger"
	"github.com/muesli/cache2go"
	"github.com/sirupsen/logrus"
	"log"
)

var (
	D *cache2go.CacheTable
)

func init () {
	D = cache2go.Cache("dkv")
	D.SetLogger(log.New(logger.D.WriterLevel(logrus.DebugLevel), "", log.LstdFlags))
}