package infrastructure

import (
	"fmt"

	"github.com/caarlos0/env"
)

type Config struct {
	MinVariance float64 `env:"MIN_VARIANCE" envDefault:"-0.2"`
	MaxVariance float64 `env:"MAX_VARIANCE" envDefault:"0.2"`
	CloneNum    int     `env:"CLONE_NUM" envDefault:"300"`

	HTTPPort               int    `env:"HTTP_PORT" envDefault:"8080"`
	HTTPInsecureSkipVerify bool   `env:"HTTP_INSECURE_SKIP_VERIFY" envDefault:"true"`
	HTTPMillisecondTimeout int    `env:"HTTP_MILLISECOND_TIMEOUT" envDefault:"1500"`
	HTTPPublishPath        string `env:"HTTP_PUBLISH_PATH" envDefault:"http://receiver-mock:8081/process"`

	PubSubSubscribeTopic     string `env:"PUB_SUB_SUBSCRIBE_TOPIC" envDefault:"device_data_test"`
	PubSubPublishTopic       string `env:"PUB_SUB_PUBLISH_TOPIC" envDefault:"device_data"`
	PubSubHandlerName        string `env:"PUB_SUB_HANDLER_NAME" envDefault:"device_data_handler"`
	PubSubSubscribeGroupName string `env:"PUB_SUB_SUBSCRIBE_GROUP_NAME" envDefault:"device_data_tester_group"`
	KafkaAddr                string `env:"KAFKA_ADDR" envDefault:"kafka:9092"`
}

func GetConfig() (*Config, error) {
	cfg := new(Config)

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return cfg, nil
}
