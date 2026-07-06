// Package password 封装密码哈希与校验（bcrypt）。
package password

import "golang.org/x/crypto/bcrypt"

// Hash 对明文密码做 bcrypt 哈希。
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare 校验明文密码与哈希是否匹配。
func Compare(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
