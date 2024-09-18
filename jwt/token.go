package jwt

import (
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/dgrijalva/jwt-go"
)

// Gen 生成Token
// Deprecated: JWT Function not safe for payload
func Gen(claims map[string]any, salt string, exp time.Duration) string {
	if exp != 0 {
		claims["exp"] = time.Now().Add(exp).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	signed, err := token.SignedString([]byte(salt))
	if err != nil {
		logs.Error("Token General Failed:", err)
	}
	return signed
}

// Parse 解析Token
// Deprecated: JWT Function not safe for payload
func Parse(otk, salt string) (claims map[string]any) {
	token, err := jwt.Parse(otk, func(token *jwt.Token) (any, error) {
		return []byte(salt), nil
	})
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				logs.Error("That's Not A Token:", err)
				return
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				logs.Error("Token Has Expired:", err)
				return
			} else {
				logs.Error("Invalid Token:", err)
				return
			}
		} else {
			logs.Error("Token Parse Failed:", err)
			return
		}
	}
	if !token.Valid {
		logs.Error("Invalid Token:", err)
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logs.Error("Parse Token Format Error:", err)
		return
	}
	return
}
