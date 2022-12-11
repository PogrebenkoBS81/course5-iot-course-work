package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"device-event-multiplier/internal/domain"
)

type HttpDeviceDataPublisher struct {
	httpClient  HTTPClient
	publishPath string
}

func NewHttpDeviceDataPublisher(httpClient HTTPClient, publishPath string) *HttpDeviceDataPublisher {
	return &HttpDeviceDataPublisher{
		httpClient:  httpClient,
		publishPath: publishPath,
	}
}

func (r *HttpDeviceDataPublisher) PublishDeviceData(deviceData []*domain.DeviceData) error {
	log.Println("Publishing cloned info to the: ", r.publishPath)

	for _, dd := range deviceData {
		bts, err := json.Marshal(dd)
		if err != nil {
			return fmt.Errorf("marshal %v: %w", dd, err)
		}

		_, err = r.httpClient.SendRequest(
			http.MethodPost,
			r.publishPath,
			http.Header{"Content-Type": {`application/json; charset=utf-8`}},
			bts,
		)
		if err != nil {
			return fmt.Errorf("send request %v: %w", dd, err)
		}
	}

	return nil
}
