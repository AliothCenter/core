package model

import "time"

type User struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	Username  string    `gorm:"column:username;type:varchar(255);unique;not null;uniqueIndex:idx_username"`
	Nickname  string    `gorm:"column:nickname;type:varchar(255);index:idx_nickname"`
	Password  string    `gorm:"column:password;type:varchar(128);not null"`
	Email     string    `gorm:"column:email;type:varchar(255);unique;not null;uniqueIndex:idx_email"`
	Phone     string    `gorm:"column:phone;type:varchar(20);unique;uniqueIndex:idx_phone"`
	Avatar    string    `gorm:"column:avatar;type:varchar(255)"`
	Role      string    `gorm:"column:role;type:varchar(20);not null"`
	Enable    bool      `gorm:"column:enable;type:boolean;not null;default:false"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

func (User) TableName() string {
	return "alioth_users"
}

type UserDTO struct {
	Username  string    `gorm:"column:username"`
	Nickname  string    `gorm:"column:nickname"`
	Password  string    `gorm:"column:password"`
	Email     string    `gorm:"column:email"`
	Phone     string    `gorm:"column:phone"`
	Avatar    string    `gorm:"column:avatar"`
	Role      string    `gorm:"column:role"`
	Enable    bool      `gorm:"column:enable"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Application struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	Name      string    `gorm:"column:name;type:varchar(255);unique;not null;uniqueIndex:idx_name"`
	Owner     string    `gorm:"column:owner;type:varchar(255);not null;index:idx_owner"`
	Avatar    string    `gorm:"column:avatar;type:varchar(255)"`
	Website   string    `gorm:"column:website;type:varchar(255)"`
	Email     string    `gorm:"column:email;type:varchar(255);not null"`
	Callback  string    `gorm:"column:callback;type:varchar(255)"`
	Enable    bool      `gorm:"column:enable;type:boolean;not null;default:false"`
	AppKey    string    `gorm:"column:app_key;type:varchar(32);unique;not null;uniqueIndex:idx_app_key"`
	AppSecret string    `gorm:"column:app_secret;type:varchar(32);unique;not null;uniqueIndex:idx_app_secret"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

func (Application) TableName() string {
	return "alioth_applications"
}

type ApplicationDTO struct {
	Name      string    `gorm:"column:name"`
	Owner     string    `gorm:"column:owner"`
	Avatar    string    `gorm:"column:avatar"`
	Website   string    `gorm:"column:website"`
	Email     string    `gorm:"column:email"`
	Callback  string    `gorm:"column:callback"`
	Enable    bool      `gorm:"column:enable"`
	AppKey    string    `gorm:"column:app_key"`
	AppSecret string    `gorm:"column:app_secret"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

type Scope struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement;not null"`
	Application string    `gorm:"column:application;type:varchar(255);not null;uniqueIndex:idx_name"`
	Name        string    `gorm:"column:name;type:varchar(255);not null;uniqueIndex:idx_name"`
	Description string    `gorm:"column:description;type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime"`
}

func (Scope) TableName() string {
	return "alioth_scopes"
}

type ScopeDTO struct {
	Name        string    `gorm:"column:name"`
	Application string    `gorm:"column:application"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}
