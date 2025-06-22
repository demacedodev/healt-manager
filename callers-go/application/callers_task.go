package application

import (
	"callers-go/domain"
	"callers-go/infrastructure/repository"
	"callers-go/pkg/async"
	"fmt"
)

type (
	App app
)

func NewApp(cfg *Config) domain.Manager {
	return &app{
		client: cfg.Client,
		db:     cfg.Storage,
		mem:    cfg.Cache,
	}
}

func (a *App) LoadDevices() error {
	tinyDevices, err := a.client.DeviceRawInformation()
	if err != nil {
		return err
	}

	fmt.Printf("☢️ [LoadDevices] Devices Network Found: %d\n", len(tinyDevices))

	deviceIDs := make([]string, 0)
	networkDevices := make([]domain.Device, 0)

	for _, tinyDevice := range tinyDevices {
		device, ok := tinyDevice.(map[string]any)
		if !ok {
			fmt.Printf("🔴 [LoadDevices] Empty Map Tiny Devices: %v\n", tinyDevices)
			continue
		}

		newDevice := domain.Device{
			DeviceId:       device["id"].(string),
			DeviceIp:       device["ip"].(string),
			DevicePassword: device["key"].(string),
			DeviceName:     device["name"].(string),
			DeviceNickName: device["name"].(string),
			Location: &domain.Location{
				Room: "R1",
				Bed:  "B1",
				Zone: "ALL",
			},
		}

		networkDevices = append(networkDevices, newDevice)
		deviceIDs = append(deviceIDs, newDevice.DeviceId)
	}

	if len(networkDevices) == 0 {
		fmt.Println("❌ [LoadDevices] Empty Storage Devices")
		return nil
	}

	dbDevices, err := a.db.GetDevices(&domain.Search{DeviceIDs: deviceIDs})
	if err != nil {
		return err
	}

	var storedDevices = make(map[string]domain.Device)
	for _, dbDevice := range dbDevices {
		storedDevices[dbDevice.DeviceId] = dbDevice
	}

	noConfiguratedDevices := make([]domain.Device, 0)
	for i := 0; i < len(networkDevices); i++ {
		auxDevice, ok := storedDevices[networkDevices[i].DeviceId]
		if !ok {
			fmt.Println("🟡 [LoadDevices] Device Not Configurated Will Be Loaded" + networkDevices[i].String())
			noConfiguratedDevices = append(noConfiguratedDevices, networkDevices[i])
			continue
		}

		networkDevices[i].DeviceNickName = auxDevice.DeviceNickName
		networkDevices[i].Location.Room = auxDevice.Location.Room
		networkDevices[i].Location.Bed = auxDevice.Location.Bed
		networkDevices[i].Location.Zone = auxDevice.Location.Zone
	}

	if err = a.db.CreateDevices(noConfiguratedDevices); err != nil {
		fmt.Printf("❌ [LoadDevices] Cannot Create Storage Devices: %v\n", err)
	}

	if err = a.mem.CreateDevices(networkDevices); err != nil {
		fmt.Printf("❌ [LoadDevices] Cannot Update Mem Devices: %v\n", err)
		return err
	}

	return nil
}

func (a *App) UpdateStatus() error {
	devices, err := a.db.GetDevices(&domain.Search{DeviceZone: repository.FULL})
	if err != nil {
		return err
	}

	update := func(device domain.Device) domain.Device {
		devicePY, errUpdate := a.client.GetDeviceStatus(&domain.Search{
			DeviceIp:   device.DeviceIp,
			DeviceId:   device.DeviceId,
			DevicePass: device.DevicePassword,
		})
		if errUpdate != nil {
			fmt.Printf("💥 [UpdateStatus] Cannot Get Device Status: %v\n", errUpdate)
			return device
		}

		device.DeviceStatus = devicePY.DeviceStatus
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
