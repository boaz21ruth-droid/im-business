package model

import "time"

type User struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	UserID         string    `gorm:"column:user_id;uniqueIndex;size:64;not null"`
	Account        *string   `gorm:"column:account;uniqueIndex;size:64"`
	PhoneNumber    *string   `gorm:"column:phone_number;uniqueIndex;size:32"`
	AreaCode       string    `gorm:"column:area_code;size:8"`
	Email          *string   `gorm:"column:email;uniqueIndex;size:128"`
	Password       string    `gorm:"column:password;size:256;not null"` // bcrypt(MD5)
	Nickname       string    `gorm:"column:nickname;size:64;not null;default:''"`
	FaceURL        string    `gorm:"column:face_url;size:512;not null;default:''"`
	Gender         int8      `gorm:"column:gender;not null;default:0"`
	Birth          int64     `gorm:"column:birth;not null;default:0"`
	Level          int8      `gorm:"column:level;not null;default:1"`
	AllowAddFriend int8      `gorm:"column:allow_add_friend;not null;default:1"`
	AllowBeep      int8      `gorm:"column:allow_beep;not null;default:1"`
	AllowVibration int8      `gorm:"column:allow_vibration;not null;default:1"`
	Status         int8      `gorm:"column:status;not null;default:1"` // 1=normal 2=banned
	TotpSecret     string    `gorm:"column:totp_secret;size:256;not null;default:''"`
	TotpEnabled    bool      `gorm:"column:totp_enabled;not null;default:false"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (User) TableName() string { return "users" }
