package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"device-event-multiplier/internal/domain"
)

type PubSubDeviceDataPublisher struct {
	publisher    message.Publisher
	publishTopic string
}

func NewPubSubDeviceDataPublisher(publisher message.Publisher, publishTopic string) *PubSubDeviceDataPublisher {
	return &PubSubDeviceDataPublisher{
		publisher:    publisher,
		publishTopic: publishTopic,
	}
}

func (r *PubSubDeviceDataPublisher) PublishDeviceData(deviceData []*domain.DeviceData) error {
	log.Println("Publishing cloned info to the pub sub topic")

	for _, dd := range deviceData {
		bts, err := json.Marshal(dd)
		if err != nil {
			return fmt.Errorf("marshal %v: %w", dd, err)
		}

		if err := r.publisher.Publish(r.publishTopic, message.NewMessage(watermill.NewUUID(), bts)); err != nil {
			return fmt.Errorf("publish %s: %w", bts, err)
		}
	}

	return nil
}
