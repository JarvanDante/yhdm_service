// Package totp 复刻旧系统的谷歌验证码（Google Authenticator / TOTP）校验。
// 旧 PHP 用自带 Plugins/GoogleAuthenticator，密钥存 admin_user.google_code。
package totp

import "github.com/pquerna/otp/totp"

// Verify 校验一次性验证码是否与密钥匹配。
func Verify(secret, code string) bool {
	if secret == "" {
		return false
	}
	return totp.Validate(code, secret)
}
