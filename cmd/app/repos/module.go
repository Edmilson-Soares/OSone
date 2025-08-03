package repos

import (
	"time"

	"gorm.io/gorm"
)

type Virtual struct {
	ID           string            `gorm:"primaryKey;type:uuid" json:"id"`
	Name         string            `gorm:"size:255;not null" json:"name"`
	Icon         string            `gorm:"size:512" json:"icon"`
	Desc         string            `gorm:"type:text" json:"desc"`
	Auth         map[string]string `gorm:"type:jsonb;serializer:json" json:"auth"`
	EnterpriseId string            `gorm:"type:uuid;not null" json:"enterpriseId"`
	Devices      []Device          `gorm:"foreignKey:VirtualId;constraint:OnDelete:CASCADE" json:"devices"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt    `gorm:"index" json:"-"`
}

type Device struct {
	ID        string                 `gorm:"primaryKey;type:uuid" json:"id"`
	Code      string                 `gorm:"size:255;uniqueIndex;not null" json:"code"`
	Name      string                 `gorm:"size:255;not null" json:"name"`
	Icon      string                 `gorm:"size:512" json:"icon"`
	Desc      string                 `gorm:"type:text" json:"desc"`
	VirtualId string                 `gorm:"type:uuid;not null" json:"virtualId"`
	Virtual   Virtual                `gorm:"foreignKey:VirtualId;references:ID" json:"virtual"`
	Location  map[string]string      `gorm:"type:jsonb;serializer:json" json:"location"`
	Config    map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"config"`
	Auth      map[string]string      `gorm:"type:jsonb;serializer:json" json:"auth"`
	Network   map[string]string      `gorm:"type:jsonb;serializer:json" json:"network"`
	CreatedAt time.Time              `json:"createdAt"`
	UpdatedAt time.Time              `json:"updatedAt"`
	DeletedAt gorm.DeletedAt         `gorm:"index" json:"-"`
}
