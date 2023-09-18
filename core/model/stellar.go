package model

import "time"

type InstancePO struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	Address   string    `gorm:"column:address;type:varchar(21);unique;not null;uniqueIndex:idx_address"`
	Name      string    `gorm:"column:name;type:varchar(255);unique;not null;uniqueIndex:idx_name"`
	Service   string    `gorm:"column:service;type:varchar(255);not null;index:idx_service"`
	Version   uint64    `gorm:"column:version;not null;index:idx_version"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

func (InstancePO) TableName() string {
	return "alioth_instances"
}

type InstanceDTO struct {
	Address   string    `gorm:"column:address"`
	Name      string    `gorm:"column:name"`
	Service   string    `gorm:"column:service"`
	Version   uint64    `gorm:"column:version"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}
