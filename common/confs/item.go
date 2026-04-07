package confs

type Item struct {
	CfgID        int32  `gorm:"column:cfg_id;index" bson:"cfg_id"`         // 配置表ID
	IsBind       int32  `gorm:"column:is_bind" bson:"is_bind"`             // 0=非绑定 1=绑定
	IsUnique     int32  `gorm:"column:is_unique" bson:"is_unique"`         // 0=非唯一 1=唯一
	ExpireType   int32  `gorm:"column:expire_type" bson:"expire_type"`     // 过期类型 0=永久 1=限时 2=限定时间
	ResourceType int32  `gorm:"column:resource_type" bson:"resource_type"` //道具类型
	LimitTime    int64  `gorm:"column:limit_time" bson:"limit_time"`       // 过期时间
	Name         string `gorm:"column:name" bson:"name"`
}
