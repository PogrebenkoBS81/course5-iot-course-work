package application

import (
	"math/rand"
	"time"

	"device-event-multiplier/internal/domain"
)

type DeviceProcessor struct {
	minVariance float64
	maxVariance float64
	cloneNum    int
}

func NewDeviceProcessor(minVariance, maxVariance float64, cloneNum int) *DeviceProcessor {
	rand.Seed(time.Now().UnixNano())

	return &DeviceProcessor{
		minVariance: minVariance,
		maxVariance: maxVariance,
		cloneNum:    cloneNum,
	}
}

func (s *DeviceProcessor) CloneDeviceData(d *domain.DeviceData) []*domain.DeviceData {
	clonedDeviceData := make([]*domain.DeviceData, 0, s.cloneNum)

	for i := 0; i < s.cloneNum; i++ {
		clonedDeviceData = append(clonedDeviceData, &domain.DeviceData{
			DeviceId:    d.DeviceId,
			Time:        d.Time + int64(i) + 1,
			Temperature: mutateFloat64(d.Temperature, s.minVariance, s.maxVariance),
			Humidity:    mutateFloat64(d.Humidity, s.minVariance, s.maxVariance),
		})
	}

	return clonedDeviceData
}

func mutateFloat64(initValue, minVariance, maxVariance float64) float64 {
	variance := getRandomFloatInRage(minVariance, maxVariance)

	return initValue + initValue*variance
}

func getRandomFloatInRage(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
