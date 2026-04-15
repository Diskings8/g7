package globals

// 订单状态
const (
	OrderStatusPending    = 1 // 待支付
	OrderStatusPaid       = 2 // 已支付
	OrderStatusProcessing = 3 // 发货中
	OrderStatusCompleted  = 4 // 已完成
	OrderStatusCancelled  = 5 // 已取消
	OrderStatusFailed     = 6 // 支付失败
	OrderStatusRefunded   = 7 // 已退款
)

// 支付方式
const (
	PayTypeWeChat = 1 // 微信支付
	PayTypeAlipay = 2 // 支付宝
	PayTypeApple  = 3 // Apple IAP
	PayTypeGoogle = 4 // Google Play
)
