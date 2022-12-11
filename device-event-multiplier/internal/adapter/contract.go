package adapter

import "device-event-multiplier/internal/domain"

type DeviceDataProcessor interface {
	CloneDeviceData(d *domain.DeviceData) []*domain.DeviceData
}

type DeviceDataPublisher interface {
	PublishDeviceData(devicesData []*domain.DeviceData) error
}
