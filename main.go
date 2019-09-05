package main

import (
	"github.com/mqiqe/prometheus-kafka-m3db/pkg/kafkaservice"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}
func main() {
	// 接收信息
	log.Infof("start : %v", time.Now())
	storeUrl := "http://10.254.192.2:7201/api/v1/prom/remote/write"
	ks := kafkaservice.NewKafkaService("localhost", "metrics", "earliest", storeUrl)
	for i := 0; i < 21; i++ {
		go ks.Consumer()
	}
	ks.Consumer()
	log.Info("start....")
}
