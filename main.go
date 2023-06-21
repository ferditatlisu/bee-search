package main

import (
	"bee-search/application/handler"
	"bee-search/pkg/client"
	"bee-search/pkg/config"
	"bee-search/pkg/model"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	select {
	case <-time.After(15 * time.Second):
		fmt.Println("Waiting to start ...")
	}
	configInstance := config.CreateConfigInstance()
	applicationConfig, _ := configInstance.GetConfig()
	kafka := getConsumer(applicationConfig)
	redis := client.NewRedisService(applicationConfig.Redis)
	searchData := &model.SearchData{
		PodName:     os.Getenv("POD_NAME"),
		Key:         os.Getenv("KEY"),
		Topic:       os.Getenv("TOPIC"),
		Value:       os.Getenv("VALUE"),
		MetadataKey: os.Getenv("METADATA_KEY"),
	}

	sh := handler.NewConsumeHandler(kafka, redis, applicationConfig, searchData)
	sh.Handle()

	fmt.Println("waiting to kill ...")
	if <-time.After(10 * time.Second); true {
		fmt.Println("still waiting ...")
	}
}

func getConsumer(config *config.ApplicationConfig) client.KafkaConsumer {
	kafkaConfig := getKafkaConfig(config)
	if kafkaConfig.UserName == nil {
		return client.NewEarthKafkaConsumer(kafkaConfig.Host)
	}

	return client.NewSecureKafkaConsumer(kafkaConfig)
}

func getKafkaConfig(config *config.ApplicationConfig) *config.KafkaConfig {
	for i := range config.KafkaConfigs {
		kafkaConfig := config.KafkaConfigs[i]
		id := strconv.Itoa(kafkaConfig.Id)
		if os.Getenv("KAFKA_ID") == id {
			return &kafkaConfig
		}
	}

	panic("Not available kafka config")
}
