package application

import (
	"callers-go/domain"
	"context"
	"errors"
	"fmt"
	"strings"
)

type (
	app struct {
		client domain.Client
		mem    domain.Repository
	}

	Config struct {
		Client domain.Client
		Cache  domain.Repository
	}

	Callers domain.Manager
)

func NewTask(cfg *Config) *App {
	return &App{
		client: cfg.Client,
		mem:    cfg.Cache,
	}
}

func (a *app) GetDevices(ctx context.Context, zone string) ([]domain.Device, error) {
	devices, err := a.client.DeviceRawInformation()
	if err != nil {
		return nil, err
	}

	_devices := make([]domain.Device, 0)

	for _, device := range devices {
		if !strings.Contains(device.EntityId, "light.") {
			continue
		}

		d := ParseDevice(device)
		if d == nil {
			fmt.Printf("🔴 [LoadDevices] cannot Decode Device: %v\n", device)
			continue
		}

		var state bool
		if strings.EqualFold(device.State, "on") {
			state = true
		}

		_devices = append(_devices, domain.Device{
			DeviceId:     device.EntityId,
			DeviceName:   d.Name,
			DeviceStatus: state,
			Location: &domain.Location{
				Room: d.Room,
				Bed:  d.Bed,
				Zone: d.Zone,
			},
		})
	}

	if len(_devices) == 0 {
		fmt.Println("❌ [LoadDevices] Empty Configured Devices")
		return nil, errors.New("devices not found")
	}

	return _devices, nil
}

func (a *app) CreateDevice(ctx context.Context, devices []domain.Device) error {
	panic("implement me")
}
