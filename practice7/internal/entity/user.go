package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID      `json:"ID" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time      `json:"CreatedAt"`
	UpdatedAt time.Time      `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `json:"DeletedAt" gorm:"index"`
	Username  string         `json:"Username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"Email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"Password"`
	Role      string         `json:"Role" gorm:"default:user"`
	Verified  bool           `json:"Verified" gorm:"default:false"`
	VerifyCode string        `json:"-" gorm:"column:verify_code"`
}
