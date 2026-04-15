package model_common

// 订单表
type GameOrder struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement"`
	OrderNo      string `gorm:"column:order_no;uniqueIndex;size:64"` // 订单号
	PlayerID     int64  `gorm:"column:player_id;index"`              // 玩家ID
	ServerID     int32  `gorm:"column:server_id;index"`              // 服务器ID
	ProductID    int32  `gorm:"column:product_id;size:64"`           // 商品ID
	ProductName  string `gorm:"column:product_name;size:128"`        // 商品名称
	ProductType  int32  `gorm:"column:product_type"`                 // 商品类型(1:道具 2:礼包 3:月卡 4:钻石)
	Price        int64  `gorm:"column:price"`                        // 价格(分)
	Currency     string `gorm:"column:currency;size:8;default:CNY"`  // 货币类型
	PayType      int32  `gorm:"column:pay_type"`                     // 支付方式
	PayAmount    int64  `gorm:"column:pay_amount"`                   // 实际支付金额(分)
	Status       int32  `gorm:"column:status;index;default:1"`       // 订单状态
	ExtInfo      string `gorm:"column:ext_info;type:text"`           // 扩展信息(JSON)
	CreateTime   int64  `gorm:"column:create_time;autoCreateTime"`   // 创建时间
	PayTime      int64  `gorm:"column:pay_time"`                     // 支付时间
	CompleteTime int64  `gorm:"column:complete_time"`                // 完成时间
	CallbackData string `gorm:"column:callback_data;type:text"`      // 支付回调原始数据
}

func (m *GameOrder) TableName() string {
	return "game_order"
}

// 订单物品明细
type OrderItem struct {
	ID        int64  `gorm:"column:id;primaryKey"`
	OrderNo   string `gorm:"column:order_no;index"`
	PlayerID  int64  `gorm:"column:player_id;index"`
	ItemID    int32  `gorm:"column:item_id"`    // 道具ID
	ItemName  string `gorm:"column:item_name"`  // 道具名称
	ItemCount int32  `gorm:"column:item_count"` // 数量
	UnitPrice int64  `gorm:"column:unit_price"` // 单价(分)
}

// 支付流水
type PaymentRecord struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	TradeNo    string `gorm:"column:trade_no;uniqueIndex;size:64"` // 第三方交易号
	OrderNo    string `gorm:"column:order_no;index;size:64"`       // 内部订单号
	PlayerID   int64  `gorm:"column:player_id;index"`
	PayType    int32  `gorm:"column:pay_type"`              // 支付方式
	Amount     int64  `gorm:"column:amount"`                // 金额(分)
	Status     int32  `gorm:"column:status"`                // 1:成功 2:失败 3:退款
	NotifyTime int64  `gorm:"column:notify_time"`           // 回调时间
	NotifyData string `gorm:"column:notify_data;type:text"` // 回调原始数据
	CreateTime int64  `gorm:"column:create_time;autoCreateTime"`
}

func (m *PaymentRecord) TableName() string {
	return "payment_record"
}
