package globals

// 邮件类型常量
const (
	MailTypeSystem = 1 // 系统邮件（全服所有玩家）
	MailTypeServer = 2 // 全服邮件（指定服务器）
	MailTypeGuild  = 3 // 全工会邮件
)

// 邮件状态常量
const (
	MailStatusPending   = 1 // 待发送
	MailStatusSending   = 2 // 发送中
	MailStatusCompleted = 3 // 已完成
	MailStatusCancelled = 4 // 已取消
)
