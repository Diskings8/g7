package req_game

type SelectPlayerReq struct {
	UserID   int64 `json:"user_id" binding:"required"`
	UID      int64 `json:"uid" binding:"required"`
	ServerID int   `json:"server_id" binding:"required"`
}

type CreatePlayerReq struct {
	UserID   int64  `json:"user_id" binding:"required"`
	Nickname string `json:"nickname" binding:"required"`
	ServerID int    `json:"server_id" binding:"required"`
}
