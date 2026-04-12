package support

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// AuthClaims 表示登录令牌里的数据。
type AuthClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	ExpireAt int64  `json:"expire_at"`
}

// GenerateToken 生成一个带签名的简单令牌。
func GenerateToken(secret string, claims AuthClaims) (string, error) {
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signature := sign(secret, payload)

	return payload + "." + signature, nil
}

// ParseToken 解析并验证令牌。
func ParseToken(secret string, token string) (AuthClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return AuthClaims{}, errors.New("token 格式不正确")
	}

	expectedSignature := sign(secret, parts[0])
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return AuthClaims{}, errors.New("token 签名无效")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return AuthClaims{}, err
	}

	var claims AuthClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return AuthClaims{}, err
	}

	if claims.ExpireAt < time.Now().Unix() {
		return AuthClaims{}, errors.New("token 已过期")
	}

	return claims, nil
}

func sign(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
