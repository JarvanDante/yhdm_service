package password

import "testing"

// 参考值由独立实现（Python md5）计算，等价于旧 PHP
// AdminUserService::makePassword，用于锁死哈希算法的兼容性。
func TestMake_MatchesLegacyPHP(t *testing.T) {
	cases := []struct {
		plain, salt, want string
	}{
		{"admin123", "12345", "940a2974608474c2289794f404b34876"},
		{"P@ssw0rd", "48213", "5df90741ff8f5fc1d93ca090456fae25"},
		{"樱花动漫", "30001", "97f97f47081d1a66899cb63ea03b7986"},
	}
	for _, c := range cases {
		if got := Make(c.plain, c.salt); got != c.want {
			t.Errorf("Make(%q,%q)=%s, want %s", c.plain, c.salt, got, c.want)
		}
		if !Verify(c.plain, c.salt, c.want) {
			t.Errorf("Verify(%q,%q,%q) should be true", c.plain, c.salt, c.want)
		}
	}
	if Verify("wrong", "12345", "940a2974608474c2289794f404b34876") {
		t.Error("Verify should fail for wrong password")
	}
}
