package client

import (
	"callers-go/domain"
	"encoding/json"
	"fmt"
	"resty.dev/v3"
	"time"
)

type (
	instance struct {
		callersBaseURL string
		devicesBaseURL string
		client         *resty.Client
	}

	Config struct {
		CallersBaseURL string
		DevicesBaseURL string
		Timeout        time.Duration
	}
)

func NewClient(cfg *Config) domain.Client {
	return &instance{
		callersBaseURL: cfg.CallersBaseURL,
		devicesBaseURL: cfg.DevicesBaseURL,
		client:         resty.New().SetTimeout(cfg.Timeout),
	}
}

func (i *instance) GetDeviceStatus(s *domain.Search) (*domain.Device, error) {
	url := fmt.Sprintf("%s/health/device/%s/status", i.callersBaseURL, s.DeviceId)

	response, err := i.client.R().
		SetQueryParam("device_ip", s.DeviceIp).
		SetQueryParam("device_password", s.DevicePass).
		Get(url)
	if err != nil {
		return nil, &domain.Error{
			Code:    "CONN-001",
			Message: err.Error(),
		}
	}

	if response.IsError() {
		return nil, &domain.Error{
			Code:        "CONN-002",
			Message:     "devices information not available",
			RawResponse: response.String(),
		}
	}

	var d *domain.Device
	if err = json.Unmarshal(response.Bytes(), &d); err != nil {
		return nil, &domain.Error{
			Code:        "CONN-003",
			Message:     err.Error(),
			RawResponse: response.String(),
		}
	}

	return d, nil
}

func (i *instance) DeviceRawInformation() (map[string]any, error) {
	url := fmt.Sprintf("%s/devices", i.devicesBaseURL)

	response, err := i.client.R().Get(url)
	if err != nil {
		return nil, &domain.Error{
			Code:    "CONN-001",
			Message: err.Error(),
		}
	}

	if response.IsError() {
		return nil, &domain.Error{
			Code:        "CONN-002",
			Message:     "devices information not available",
			RawResponse: response.String(),
		}
	}

	var r map[string]any
	if err = json.Unmarshal(response.Bytes(), &r); err != nil {
		return nil, &domain.Error{
			Code:        "CONN-003",
			Message:     err.Error(),
			RawResponse: response.String(),
		}
	}

	return r, nil
}
