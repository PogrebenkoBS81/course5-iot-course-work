package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"device-event-multiplier/internal/domain"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type PubSubAdapter struct {
	router              *message.Router
	deviceDataProcessor DeviceDataProcessor
}

func NewPubSubAdapter(
	deviceDataProcessor DeviceDataProcessor,
	publisher message.Publisher,
	subscriber message.Subscriber,
	handlerName string,
	subscribeTopic string,
	publishTopic string,
) (*PubSubAdapter, error) {
	router, err := message.NewRouter(message.RouterConfig{}, watermill.NewStdLogger(false, false))
	if err != nil {
		return nil, fmt.Errorf("new router %w", err)
	}

	s := &PubSubAdapter{
		router:              router,
		deviceDataProcessor: deviceDataProcessor,
	}

	router.AddHandler(
		handlerName,
		subscribeTopic,
		subscriber,
		publishTopic,
		publisher,
		s.processDeviceDataCloning,
	)

	return s, nil
}

func (s *PubSubAdapter) processDeviceDataCloning(msg *message.Message) ([]*message.Message, error) {
	unprocessed := new(domain.DeviceData)

	if err := json.Unmarshal(msg.Payload, unprocessed); err != nil {
		return nil, fmt.Errorf("unmarshal %q: %w", msg.Payload, err)
	}

	processedData := s.deviceDataProcessor.CloneDeviceData(unprocessed)
	processedMessages := make([]*message.Message, 0, len(processedData))

	for _, processedMessage := range processedData {
		bts, err := json.Marshal(processedMessage)
		if err != nil {
			return nil, fmt.Errorf("marshal %v: %w", processedData, err)
		}

		processedMessages = append(processedMessages, message.NewMessage(watermill.NewUUID(), bts))
	}

	return processedMessages, nil
}

func (s *PubSubAdapter) Run(ctx context.Context) error {
	if err := s.router.Run(ctx); err != nil {
		return fmt.Errorf("run router %w", err)
	}

	<-ctx.Done()

	return nil
}
