package model_common

import (
	"database/sql/driver"
	"encoding/json"
	"g7/common/globals"
	"gorm.io/gorm"
	"time"
)

// PlayerMail 玩家邮件表
type PlayerMail struct {
	gorm.Model         // 自带 ID, CreatedAt, UpdatedAt, DeletedAt (软删除)
	PlayerID    uint64 `gorm:"column:player_id;index;not null;comment:玩家ID"`
	ServerID    uint64 `gorm:"column:server_id;not null;comment:服务器ID"`
	Title       string `gorm:"column:title;size:128;not null;comment:邮件标题"`
	Content     string `gorm:"column:content;type:text;comment:邮件内容"`
	MailType    int32  `gorm:"column:mail_type;default:0;comment:邮件类型 0=普通 1=系统 2=活动 3=奖励"`
	Status      int32  `gorm:"column:status;default:0;index;comment:0未读 1已读 2已删除"`
	HasAttach   int32  `gorm:"column:has_attach;default:0;comment:0无附件 1有附件"`
	AttachItems string `gorm:"column:attach_items;type:text;comment:附件道具JSON"`
	ExpireAt    int64  `gorm:"column:expire_at;index;comment:过期时间"`
	SendFrom    string `gorm:"column:send_from;size:64;default:system;comment:发送方"`
}

// TableName 自定义表名
func (PlayerMail) TableName() string {
	return "player_mail"
}

// Attachment 附件结构
type Attachment struct {
	ItemID int64 `json:"item_id"`
	Count  int32 `json:"count"`
}

// Attachments 附件列表，实现 Scanner/Valuer 接口
type Attachments []Attachment

func (a Attachments) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

func (a *Attachments) Scan(value interface{}) error {
	if value == nil {
		*a = Attachments{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// BaseMail 基础邮件表（模板表）
type BaseMail struct {
	ID          int64       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MailType    int8        `gorm:"column:mail_type;type:tinyint;not null;default:1;comment:邮件类型: 1=系统邮件 2=全服邮件 3=全工会邮件" json:"mail_type"`
	Title       string      `gorm:"column:title;type:varchar(128);not null;comment:邮件标题" json:"title"`
	Content     string      `gorm:"column:content;type:text;not null;comment:邮件内容" json:"content"`
	Attachments Attachments `gorm:"column:attachments;type:json;comment:附件列表" json:"attachments"`

	// 发送范围
	TargetServerID *int64 `gorm:"column:target_server_id;type:bigint;default:null;comment:目标服务器ID(全服邮件可为NULL)" json:"target_server_id,omitempty"`
	TargetGuildID  *int64 `gorm:"column:target_guild_id;type:bigint;default:null;comment:目标工会ID(工会邮件使用)" json:"target_guild_id,omitempty"`

	// 时间控制
	StartTime time.Time `gorm:"column:start_time;type:datetime;not null;comment:开始发送时间" json:"start_time"`
	EndTime   time.Time `gorm:"column:end_time;type:datetime;not null;comment:过期时间" json:"end_time"`

	// 状态
	Status int8 `gorm:"column:status;type:tinyint;not null;default:1;comment:1=待发送 2=发送中 3=已完成 4=已取消" json:"status"`

	// 统计
	TotalTargetCount int32 `gorm:"column:total_target_count;type:int;not null;default:0;comment:目标玩家总数" json:"total_target_count"`
	SentCount        int32 `gorm:"column:sent_count;type:int;not null;default:0;comment:已发送数量" json:"sent_count"`

	// 创建信息
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	CreatedBy string    `gorm:"column:created_by;type:varchar(64);default:null;comment:创建人" json:"created_by"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

// TableName 指定表名
func (BaseMail) TableName() string {
	return "base_mail"
}

// IsExpired 判断邮件是否已过期
func (m *BaseMail) IsExpired() bool {
	return time.Now().After(m.EndTime)
}

// IsValid 判断邮件是否有效（在有效期内且状态正常）
func (m *BaseMail) IsValid() bool {
	now := time.Now()
	return now.After(m.StartTime) && now.Before(m.EndTime) && m.Status == globals.MailStatusCompleted
}

// CanSend 判断是否可以发送
func (m *BaseMail) CanSend() bool {
	return m.Status == globals.MailStatusPending && time.Now().After(m.StartTime)
}
