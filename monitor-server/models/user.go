package models

import "time"

type User struct {
	Id       int    `json:"id" xorm:"id"`
	UserName string `json:"name" xorm:"name"`
	Password string `json:"pwd" xorm:"pwd"`
}

type UpdateUserDto struct {
	NewPassword   string `form:"new_password" json:"new_password"`
	ReNewPassword string `form:"re_new_password" json:"re_new_password"`
	DisplayName   string `form:"display_name" json:"display_name"`
	Email         string `form:"email" json:"email"`
	Phone         string `form:"phone" json:"phone"`
}

type Session struct {
	User   string `json:"user"`
	Token  string `json:"token"`
	Expire int64  `json:"expire"`
}

type UserTable struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Passwd        string    `json:"passwd"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	ExtContactOne string    `json:"ext_contact_one"`
	ExtContactTwo string    `json:"ext_contact_two"`
	Created       time.Time `json:"created"`
}

type UserQuery struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Passwd        string    `json:"passwd"`
	DisplayName   string    `json:"display_name"`
	Role          string    `json:"role"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Creator       string    `json:"creator"`
	Created       time.Time `json:"created"`
	CreatedString string    `json:"created_string"`
}

type RoleTable struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	Parent        int       `json:"parent"`
	Creator       string    `json:"creator"`
	Created       time.Time `json:"created"`
	MainDashboard int       `json:"main_dashboard"`
	Disable       int       `json:"disable"`
	OwnerUser     string    `json:"owner_user"`
}

type RoleQuery struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"display_name"`
	Email         string    `json:"email"`
	Creator       string    `json:"creator"`
	Created       time.Time `json:"created"`
	CreatedString string    `json:"created_string"`
}

type SendAlertObj struct {
	Accept  []string
	Subject string
	Content string
}

type UpdateRoleDto struct {
	RoleId      int    `json:"role_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Parent      int    `json:"parent"`
	Operator    string `json:"operator"`
	Operation   string `json:"operation" binding:"required"`
}

type RelRoleUserTable struct {
	Id     int `json:"id"`
	RoleId int `json:"role_id"`
	UserId int `json:"user_id"`
}

type UpdateRoleUserDto struct {
	RoleId int   `json:"role_id" binding:"required"`
	UserId []int `json:"user_id"`
}

type RelRoleGrpTable struct {
	Id     int `json:"id"`
	RoleId int `json:"role_id"`
	GrpId  int `json:"grp_id"`
}

type RoleGrpDto struct {
	GrpId    string `json:"grp_id"`
	GrpIdInt int    `json:"-"`
	RoleId   []int  `json:"role_id"`
}

type CoreRoleDto struct {
	Status  string            `json:"status"`
	Message string            `json:"message"`
	Data    []CoreRoleDataObj `json:"data"`
}

type CoreRoleDataObj struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type CoreRequestToken struct {
	Sub        string `json:"sub"`
	Iat        int64  `json:"iat"`
	Type       string `json:"type"`
	ClientType string `json:"clientType"`
	Exp        int64  `json:"exp"`
	Authority  string `json:"authority"`
}

type CoreJwtToken struct {
	User   string   `json:"user"`
	Expire int64    `json:"expire"`
	Roles  []string `json:"roles"`
}

type RequestCoreVariableDto struct {
	Filters  []*CoreVariableFilter `json:"filters"`
	Pageable CoreVariablePage      `json:"pageable"`
	Paging   bool                  `json:"paging"`
}

type CoreVariableFilter struct {
	Name     string `json:"name"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type CoreVariablePage struct {
	PageSize   int `json:"pageSize"`
	StartIndex int `json:"startIndex"`
}

type RequestCoreVariableResult struct {
	Data    CoreVariableResultData `json:"data"`
	Message string                 `json:"message"`
	Status  string                 `json:"status"`
}

type CoreVariableResultData struct {
	Contents []*CoreVariableResultContentObj `json:"contents"`
}

type CoreVariableResultContentObj struct {
	DefaultValue string `json:"defaultValue"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	PackageName  string `json:"packageName"`
	Scope        string `json:"scope"`
	Source       string `json:"source"`
	Status       string `json:"status"`
	Value        string `json:"value"`
}
