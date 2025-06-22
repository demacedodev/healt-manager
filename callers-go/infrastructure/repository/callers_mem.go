package repository

import (
	"callers-go/domain"
	"sync"
)

const (
	FULL = "FULL"
)

type (
	Memory struct {
		devices      []domain.Device
		cache        map[string][]domain.Device
		mu           *sync.RWMutex
		shouldUpdate bool
	}
)

func NewMemoryStorage() domain.Repository {
	return &Memory{
		devices: make([]domain.Device, 0),
		cache:   make(map[string][]domain.Device),
		mu:      &sync.RWMutex{},
	}
}

func (m *Memory) GetDevices(search *domain.Search) ([]domain.Device, error) {
	if !m.shouldUpdate {
		if search.DeviceZone == FULL {
			return m.devices, nil
		}

		devices, ok := m.cache[search.DeviceZone]
		if !ok {
			return []domain.Device{}, nil
		}

		return devices, nil
	}

	m.cache = make(map[string][]domain.Device)

	for _, device := range m.devices {
		devices, ok := m.cache[device.Location.Zone]
		if !ok {
			m.cache[device.Location.Zone] = append(m.cache[device.Location.Zone], device)
			continue
		}

		devices = append(devices, device)
		m.cache[device.Location.Zone] = devices
	}

	if devices, ok := m.cache[search.DeviceZone]; ok {
		return devices, nil
	}

	return []domain.Device{}, nil
}

func (m *Memory) CreateDevices(devices []domain.Device) error {
	m.mu.RLock()
	m.devices = devices
	m.shouldUpdate = true

	// just to preload cache
	_, _ = m.GetDevices(&domain.Search{})
	m.shouldUpdate = false
	m.mu.RUnlock()

	return nil
}
