package kafkaservice

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"io"
	"io/ioutil"
	"net/http"
)

const maxErrMsgLen = 256

var userAgent = fmt.Sprintf("Prometheus/%s", version.Version)

type KafkaService struct {
	servers         string
	groupId         string
	autoOffsetReset string
	storeUrl        string
}

// 创建服务
func NewKafkaService(servers, groupId, autoOffsetReset, storeUrl string) *KafkaService {
	return &KafkaService{
		servers:         servers,
		groupId:         groupId,
		autoOffsetReset: autoOffsetReset,
		storeUrl:        storeUrl,
	}
}

func (k *KafkaService) Consumer() error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":       k.servers,         //"localhost",
		"group.id":                k.groupId,         //"myGroup",
		"auto.offset.reset":       k.autoOffsetReset, //"earliest",
		"enable.auto.commit":      true,
		"auto.commit.interval.ms": 1000,
	})
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Subscribe("metrics", nil); err != nil {
		return err
	}
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			log.Infof("TopicPartition: %s \n", msg.TopicPartition)
			if err := Store(msg.Value, k.storeUrl); err != nil {
				log.Infof("store error：%v", err.Error())
			}
		} else {
			// The client will automatically try to recover from all errors.
			log.Infof("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func (k *KafkaService) Producer() error {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": k.servers})
	if err != nil {
		return err
	}
	defer p.Close()
	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Infof("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Infof("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "myTopic"
	for _, word := range []string{"Welcome", "to", "the", "Confluent", "Kafka", "Golang", "client"} {
		_ = p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          []byte(word),
		}, nil)
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
	return nil
}

// Store sends a batch of samples to the HTTP endpoint, the request is the proto marshalled
// and encoded bytes from codec.go.
func Store(req []byte, url string) error {

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(req))
	if err != nil {
		// Errors from NewRequest are from unparseable URLs, so are not
		// recoverable.
		return err
	}
	httpReq.Header.Add("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", userAgent)
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	//httpReq = httpReq.WithContext(ctx)

	//ctx, cancel := context.WithTimeout(context.Background(),10000)
	//defer cancel()

	httpClient, err := config.NewClientFromConfig(config.HTTPClientConfig{}, "remote_storage", false)
	if err != nil {
		return err
	}
	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		// Errors from client.Do are from (for example) network errors, so are
		// recoverable.
		return err
	}
	defer func() {
		_, _ = io.Copy(ioutil.Discard, httpResp.Body)
		_ = httpResp.Body.Close()
	}()

	if httpResp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = errors.Errorf("server returned HTTP status %s: %s", httpResp.Status, line)
	}
	if httpResp.StatusCode/100 == 5 {
		return err
	}
	return err
}
