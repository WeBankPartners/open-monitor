package models

import "time"

type RemoteWriteConfigTable struct {
	Id         string    `json:"id" xorm:"id" binding:"required"`
	Address    string    `json:"address" xorm:"address" binding:"required"`
	CreateAt   time.Time `json:"create_at" xorm:"create_at"`
	UpdateAt   time.Time `json:"update_at" xorm:"update_at"`
	CreateUser string    `json:"create_user" xorm:"create_user"`
	UpdateUser string    `json:"update_user" xorm:"update_user"`
	CreateTime string    `json:"create_time" xorm:"-"`
	UpdateTime string    `json:"update_time" xorm:"-"`
}
