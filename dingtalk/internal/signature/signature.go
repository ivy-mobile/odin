package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
)

// Sign 返回钉钉机器人 HMAC-SHA256 签名，结果未做 URL 转义
func Sign(timestamp int64, secret string) string {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	hash := hmac.New(sha256.New, []byte(secret))
	_, _ = hash.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
