package model_login

type PaymentCallBackReq struct {
	OrderId   string `json:"order_id"`
	PaymentId string `json:"payment_id"`
	Md5       string `json:"md5"`
	PayAmount int64  `json:"pay_amount"` // 实际支付金额(分)
	PayTime   int64  `json:"pay_time"`
	PayType   int32  `json:"pay_type"`
	Currency  string `json:"currency"`
	State     string `json:"state"`
}

type PaymentCallBackRsp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}
