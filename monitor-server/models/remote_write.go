package models

import "time"

type RemoteWriteConfigTable struct {
	Id       string    `json:"id" binding:"required"`
	Address  string    `json:"address" binding:"required"`
	CreateAt time.Time `json:"create_at"`
	UpdateAt time.Time `json:"update_at"`
}
