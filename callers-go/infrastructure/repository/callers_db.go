package repository

import (
	"callers-go/domain"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type (
	Storage struct {
		db *gorm.DB
	}

	Config struct {
		User     string
		Password string
		Host     string
		Port     string
		Database string
	}
)

func NewPersistentStorage(cfg *Config) (domain.Repository, error) {
	db, err := InitStorage(cfg)
	if err != nil {
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) GetDevices(search *domain.Search) ([]domain.Device, error) {
	builder := s.db

	if len(search.DeviceIp) > 0 {
		builder.Where("device_ip = ?", search.DeviceIp)
	}

	if len(search.DeviceId) > 0 {
		builder.Where("device_id = ?", search.DeviceId)
	}

	if len(search.DeviceZone) > 0 {
		builder.Where("device_zone = ?", search.DeviceZone)
	}

	if len(search.DeviceIDs) > 0 {
		builder.Where("device_id IN ?", search.DeviceIDs)
	}

	var devices []domain.Device
	if err := builder.Find(&devices).Error; err != nil {
		if errors.Is(gorm.ErrRecordNotFound, err) {
			return devices, nil
		}
		return nil, err
	}

	return devices, nil
}

func (s *Storage) CreateDevices(devices []domain.Device) error {
	if len(devices) == 0 {
		return nil
	}

	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "device_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"device_password", "device_status"}),
	}).Create(&devices).Error
}

func InitStorage(cfg *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.Exec("CREATE DATABASE IF NOT EXISTS callers CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error; err != nil {
		return nil, err
	}

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/callers?charset=utf8mb4&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&domain.Device{}); err != nil {
		return nil, err
	}

	return db, nil
}
