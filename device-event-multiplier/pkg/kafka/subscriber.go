package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

func NewSubscriber(kafkaAddress, group string) (message.Subscriber, error) {
	subscriberConfig := kafka.DefaultSaramaSubscriberConfig()
	subscriberConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	subscriber, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:               []string{kafkaAddress},
			Unmarshaler:           kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: subscriberConfig,
			ConsumerGroup:         group,
		},
		watermill.NewStdLogger(false, false),
	)
	if err != nil {
		return nil, fmt.Errorf("new subscriber: %w", err)
	}

	return subscriber, nil
}
