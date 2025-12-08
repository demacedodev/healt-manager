package application

import (
	"callers-go/domain"
	"callers-go/infrastructure/repository"
	"callers-go/pkg/async"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type (
	App app
)

func NewApp(cfg *Config) domain.Manager {
	return &app{
		client: cfg.Client,
		mem:    cfg.Cache,
	}
}

func (a *App) LoadDevices() error {
	devices, err := a.client.DeviceRawInformation()
	if err != nil {
		return err
	}

	fmt.Printf("☢️ [LoadDevices] Devices Network Found: %d\n", len(devices))

	foundDevices := make([]domain.Device, 0)

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

		foundDevices = append(foundDevices, domain.Device{
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

	if len(foundDevices) == 0 {
		fmt.Println("❌ [LoadDevices] Empty Configured Devices")
		return nil
	}

	if err = a.mem.CreateDevices(foundDevices); err != nil {
		fmt.Printf("❌ [LoadDevices] Cannot Update Mem Devices: %v\n", err)
		return err
	}

	return nil
}

func (a *App) UpdateStatus() error {
	devices, err := a.mem.GetDevices(&domain.Search{DeviceZone: repository.FULL})
	if err != nil {
		return err
	}

	if len(devices) == 0 {
		return nil
	}

	update := func(device domain.Device) domain.Device {
		status, errUpdate := a.client.GetDeviceStatus(&domain.Search{
			DeviceIp:   device.DeviceIp,
			DeviceId:   device.DeviceId,
			DevicePass: device.DevicePassword,
		})
		if errUpdate != nil {
			fmt.Printf("💥 [UpdateStatus] Cannot Get Device Status: %v\n", errUpdate)
			return device
		}

		fmt.Printf("✅ [UpdateStatus] Device IP: %s - ID: %s - Name:%s - Status: %v\n", device.DeviceIp, device.DeviceId, device.DeviceName, status)
		device.DeviceStatus = status
		return device
	}

	worker := async.NewWorkerPool(len(devices), len(devices))

	for _, device := range devices {
		err = worker.Submit(func() (interface{}, error) {
			return update(device), nil
		})

		if err != nil {
			return err
		}
	}

	if len(devices) == 0 {
		worker.Close()
		return nil
	}

	var updatedDevices = make([]domain.Device, 0)
	i := 0
	enableds := 0
	disableds := 0
	for result := range worker.Results() {
		device := result.Value.(domain.Device)
		updatedDevices = append(updatedDevices, device)
		if device.DeviceStatus {
			enableds++
		} else {
			disableds++
		}

		if i == len(devices)-1 {
			worker.Close()
		}
		i++
	}

	fmt.Printf("☢️ [UpdateStatus] Devices Storage Found: %d [enabled: %d] [disabled: %d]\n", len(devices), enableds, disableds)

	if err = a.mem.CreateDevices(updatedDevices); err != nil {
		return err
	}

	return nil
}

func ParseDevice(device domain.RawDevice) *domain.HaDevice {
	decoded, err := base64.StdEncoding.DecodeString(device.Attributes.FriendlyName)
	if err != nil {
		return nil
	}

	var d *domain.HaDevice
	err = json.Unmarshal(decoded, &d)
	if err != nil {
		return nil
	}

	return d
}
