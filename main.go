package main

import (
	"flag"
	"github.com/mqiqe/prometheus-kafka-m3db/pkg/kafkaservice"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

var (
	servers         = "localhost"
	groupId         = "metrics"
	autoOffsetReset = "earliest"
	storeUrl        = "http://10.254.192.2:7201/api/v1/prom/remote/write"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	flag.StringVar(&servers, "servers", "127.0.0.1", "Kafka Address")
	flag.StringVar(&groupId, "group-id", "metrics", "Kafka Group Id")
	flag.StringVar(&autoOffsetReset, "auto-offset-reset", "earliest", "kafka Auto Offset Reset")
	flag.StringVar(&storeUrl, "store-url", "http://127.0.0.1:7201/api/v1/prom/remote/write", "M3db Store Url")
	flag.Parse()
}
func main() {
	// 接收信息
	log.Infof("start : %v", time.Now())

	ks := kafkaservice.NewKafkaService(servers, groupId, autoOffsetReset, storeUrl)
	ks.Consumer()

	log.Info("start....")
}
