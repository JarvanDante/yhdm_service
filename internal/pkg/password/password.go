// Package password 复刻旧 PHP 系统的密码哈希算法，保证现有 admin_user
// 账号在 Go 侧能直接登录（双跑期间共用同一份 Mongo 数据）。
//
// 旧 PHP: AdminUserService::makePassword($password, $slat)
//   return md5(md5($password) . 'This is password' . md5($slat));
package password

import (
	"crypto/md5"
	"encoding/hex"
)

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

// Make 生成与旧系统完全一致的密码哈希。
func Make(plain, salt string) string {
	return md5hex(md5hex(plain) + "This is password" + md5hex(salt))
}

// Verify 校验明文密码是否匹配库中存储的哈希。
func Verify(plain, salt, hashed string) bool {
	return Make(plain, salt) == hashed
}
