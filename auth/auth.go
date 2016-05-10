package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"moduleab_agent/common"
)

func GetSignature(date, url string) string {
	h := hmac.New(
		sha1.New,
		[]byte(common.LoginKey),
	)
	h.Write(
		[]byte(
			fmt.Sprintf("%s%s", date, url),
		),
	)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
