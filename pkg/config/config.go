package config

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

//go:generate mockery --name=Config --output=../../../mocks/configmock
type Config interface {
	GetConfig() (*ApplicationConfig, error)
}

type ServerConfig struct {
	Port string
}

type ApplicationConfig struct {
	Server       ServerConfig
	Redis        RedisConfig
	Master       Master
	KafkaConfigs []KafkaConfig
}

type Master struct {
	Url string
}

type KafkaConfig struct {
	Id          int      `json:"id"`
	Name        string   `json:"name"`
	Host        string   `json:"host"`
	UserName    *string  `json:"userName"`
	Password    *string  `json:"password"`
	Certificate []string `json:"certificate"`
}

type RedisConfig struct {
	MasterName string
	Host       []string
	Password   string
	Database   int
}

type config struct{}

func (c *config) GetConfig() (*ApplicationConfig, error) {
	configuration := ApplicationConfig{}

	kafkaValues := os.Getenv("KAFKA_CONFIGS")
	var kafkaConfigs []KafkaConfig
	_ = json.Unmarshal([]byte(kafkaValues), &kafkaConfigs)
	configuration.KafkaConfigs = kafkaConfigs

	configuration.Master.Url = os.Getenv("POD_MASTER_URL")

	configuration.Redis.Host = strings.Split(os.Getenv("REDIS_HOST"), ",")
	configuration.Redis.Password = os.Getenv("REDIS_PASSWORD")
	configuration.Redis.MasterName = os.Getenv("REDIS_MASTERNAME")
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	configuration.Redis.Database = db

	return &configuration, nil
}

func CreateConfigInstance() *config {
	return &config{}
}
