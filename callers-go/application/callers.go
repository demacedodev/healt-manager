package application

import (
	"callers-go/domain"
	"context"
)

type (
	app struct {
		client domain.Client
		mem    domain.Repository
		db     domain.Repository
	}

	Config struct {
		Client  domain.Client
		Storage domain.Repository
		Cache   domain.Repository
	}

	Callers domain.Manager
)

func NewTask(cfg *Config) *App {
	return &App{
		client: cfg.Client,
		db:     cfg.Storage,
		mem:    cfg.Cache,
	}
}

func (a *app) GetDevices(ctx context.Context, zone string) ([]domain.Device, error) {
	devices, err := a.mem.GetDevices(&domain.Search{DeviceZone: zone})
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (a *app) CreateDevice(ctx context.Context, devices []domain.Device) error {
	panic("implement me")
}
