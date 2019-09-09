package models

import "time"

type User struct {
	Id   int     `json:"id" xorm:"id"`
	UserName string  `json:"name" xorm:"name"`
	Password  string  `json:"pwd" xorm:"pwd"`
}

type Session struct {
	User  string  `json:"user"`
	Token  string  `json:"token"`
	Expire  int64  `json:"expire"`
}

type UserTable struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
	Passwd  string  `json:"passwd"`
	Cnname  string  `json:"cnname"`
	Email  string  `json:"email"`
	Phone  string  `json:"phone"`
}

type UserInfo struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
	CnName  string  `json:"cn_name"`
}

type TeamTable struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
	Creator  int  `json:"creator"`
	Created  time.Time  `json:"created"`
}

type TeamUserDto struct {
	Id  int  `json:"id"`
	Name  string  `json:"name"`
	Created  time.Time  `json:"created"`
	Member  []*UserInfo  `json:"member"`
}

type TeamUserQuery struct {
	TId  int  `json:"t_id"`
	TName  string  `json:"t_name"`
	Created  time.Time  `json:"created"`
	UId  int  `json:"u_id"`
	UName  string  `json:"u_name"`
	Cnname  string  `json:"cnname"`
}

type TeamUpdateDto struct {
	Id  int  `form:"id" json:"id"`
	Name  string  `form:"name" json:"name"`
	Member  []int  `form:"member" json:"member"`
}