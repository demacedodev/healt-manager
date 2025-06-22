package domain

import (
	"context"
	"encoding/json"
)

type (
	Caller map[string]any

	Manager interface {
		GetDevices(ctx context.Context, zone string) ([]Device, error)
		CreateDevice(ctx context.Context, devices []Device) error
	}

	Client interface {
		GetDeviceStatus(*Search) (*Device, error)
		DeviceRawInformation() (map[string]any, error)
	}

	Repository interface {
		GetDevices(*Search) ([]Device, error)
		CreateDevices([]Device) error
	}

	Search struct {
		DeviceIp   string
		DeviceId   string
		DevicePass string
		DeviceZone string
		DeviceIDs  []string
	}

	Device struct {
		DeviceId       string         `json:"device_id" gorm:"primaryKey"`
		DeviceIp       string         `json:"device_ip"`
		DeviceName     string         `json:"device_name"`
		DeviceNickName string         `json:"device_nick_name"`
		DevicePassword string         `json:"device_password"`
		DeviceStatus   bool           `json:"device_status"`
		RawResponse    map[string]any `json:"raw_response,omitempty" gorm:"-"`
		Location       *Location      `json:"location" gorm:"embedded"`
		Error          *Error         `json:"error" gorm:"-"`
	}

	Location struct {
		Room string `json:"room"`
		Bed  string `json:"bed"`
		Zone string `json:"zone"`
	}

	Error struct {
		Code        string `json:"code"`
		Message     string `json:"message"`
		RawResponse string `json:"raw_response,omitempty"`
	}
)

func (e Error) Error() string {
	raw, _ := json.Marshal(e)
	return string(raw)
}

func (d *Device) String() string {
	raw, _ := json.Marshal(d)
	return string(raw)
}
