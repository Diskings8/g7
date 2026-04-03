package jwt

import (
	"g7/common/config"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type Claims struct {
	UserID   int64 `json:"user_id"`   // 账号ID
	UID      int64 `json:"uid"`       // 角色雪花ID
	ServerID int32 `json:"server_id"` // 游戏服ID
	jwt.RegisteredClaims
}

// GenLoginToken 登录服签发：仅含账号ID
func GenLoginToken(userID int64) (string, error) {
	expire := time.Duration(config.GCfg.JWT.ExpireHours) * time.Hour
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GCfg.JWT.Secret))
}

// GenGameToken 选角后签发：含账号ID、角色UID、服务器ID
func GenGameToken(userID, uid int64, serverID int32) (string, error) {
	expire := time.Duration(config.GCfg.JWT.ExpireHours) * time.Hour
	claims := Claims{
		UserID:   userID,
		UID:      uid,
		ServerID: serverID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GCfg.JWT.Secret))
}

// ParseToken 通用解析
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GCfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
