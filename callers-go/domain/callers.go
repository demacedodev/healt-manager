package domain

import (
	"context"
	"encoding/json"
	"time"
)

type (
	Caller map[string]any

	Manager interface {
		GetDevices(ctx context.Context, zone string) ([]Device, error)
		CreateDevice(ctx context.Context, devices []Device) error
	}

	Client interface {
		GetDeviceStatus(*Search) (bool, error)
		DeviceRawInformation() ([]RawDevice, error)
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

	DeviceStatus struct {
		DPS map[string]json.RawMessage `json:"dps"`
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

	RawDevice struct {
		EntityId   string `json:"entity_id"`
		State      string `json:"state"`
		Attributes struct {
			SupportedColorModes []string    `json:"supported_color_modes"`
			ColorMode           interface{} `json:"color_mode"`
			OffWithTransition   bool        `json:"off_with_transition"`
			OffBrightness       interface{} `json:"off_brightness"`
			Icon                string      `json:"icon"`
			FriendlyName        string      `json:"friendly_name"`
			SupportedFeatures   int         `json:"supported_features"`
		} `json:"attributes"`
		LastChanged  time.Time `json:"last_changed"`
		LastReported time.Time `json:"last_reported"`
		LastUpdated  time.Time `json:"last_updated"`
		Context      struct {
			Id       string      `json:"id"`
			ParentId interface{} `json:"parent_id"`
			UserId   interface{} `json:"user_id"`
		} `json:"context"`
	}

	HaDevice struct {
		Name string `json:"name"`
		Zone string `json:"zone"`
		Room string `json:"room"`
		Bed  string `json:"bed"`
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
