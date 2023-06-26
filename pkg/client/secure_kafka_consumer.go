package client

import (
	"bee-search/pkg/config"
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	"os"
	"strings"
	"time"
)

type secureKafkaConsumer struct {
	c *config.KafkaConfig
}

func NewSecureKafkaConsumer(c *config.KafkaConfig) KafkaConsumer {
	return &secureKafkaConsumer{c}
}

func (k *secureKafkaConsumer) CreateConsumer(topic string, partition int) *kafka.Reader {
	readerConfig := kafka.ReaderConfig{
		Brokers:   strings.Split(k.c.Host, ","),
		Partition: partition,
		Topic:     topic,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		MaxWait:   time.Second * 1,
		Dialer:    createKafkaDialer(k.c, createTLSConfig(k.c)),
	}

	newConsumer := kafka.NewReader(readerConfig)
	return newConsumer
}

func (k *secureKafkaConsumer) ReadMessage(consumer *kafka.Reader) (kafka.Message, error) {
	msg, err := consumer.ReadMessage(context.Background())
	return msg, err
}

func (k *secureKafkaConsumer) Stop(consumer *kafka.Reader) {
	_ = consumer.Close()
}

func (k *secureKafkaConsumer) GetPartitions(topicName string) ([]kafka.Partition, error) {
	server := strings.Split(k.c.Host, ",")[0]
	dialer := createKafkaDialer(k.c, createTLSConfig(k.c))
	partitions, err := dialer.LookupPartitions(context.Background(), "tcp", server, topicName)
	return partitions, err
}

func createKafkaDialer(kafkaConfig *config.KafkaConfig, tlsConfig *tls.Config) *kafka.Dialer {
	mechanism, err := scram.Mechanism(scram.SHA512, *kafkaConfig.UserName, *kafkaConfig.Password)
	if err != nil {
		panic("Error while creating SCRAM configuration, error: " + err.Error())
	}

	return &kafka.Dialer{
		TLS:           tlsConfig,
		SASLMechanism: mechanism,
	}
}

func createTLSConfig(kafkaConfig *config.KafkaConfig) *tls.Config {
	path := "resources/rootCA.pem"
	err := createFile(kafkaConfig.Certificate, path)
	if err != nil {
		panic(err.Error())
	}
	rootCA, err := os.ReadFile(path)
	if err != nil {
		panic("Error while reading Root CA file: " + path + " error: " + err.Error())
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(rootCA)

	return &tls.Config{
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}
}

func createFile(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
