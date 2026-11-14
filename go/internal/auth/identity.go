package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/pbkdf2"
	"strings"
)

var ErrInvalidPasswordHash = errors.New("invalid identity password hash")

func NormalizeEmail(v string) string { return strings.ToUpper(strings.TrimSpace(v)) }
func HashIdentityV3(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	iter := 100000
	sub := pbkdf2.Key([]byte(password), salt, iter, 32, sha512.New)
	b := make([]byte, 13+len(salt)+len(sub))
	b[0] = 1
	binary.BigEndian.PutUint32(b[1:5], 2)
	binary.BigEndian.PutUint32(b[5:9], uint32(iter))
	binary.BigEndian.PutUint32(b[9:13], uint32(len(salt)))
	copy(b[13:], salt)
	copy(b[13+len(salt):], sub)
	return base64.StdEncoding.EncodeToString(b), nil
}
func VerifyIdentityV3(encoded, password string) bool {
	b, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(b) < 13 || b[0] != 1 {
		return false
	}
	prf := binary.BigEndian.Uint32(b[1:5])
	iter := binary.BigEndian.Uint32(b[5:9])
	sl := binary.BigEndian.Uint32(b[9:13])
	if prf != 2 || iter == 0 || sl == 0 || int(13+sl) > len(b) {
		return false
	}
	salt := b[13 : 13+sl]
	want := b[13+sl:]
	got := pbkdf2.Key([]byte(password), salt, int(iter), len(want), sha512.New)
	return hmac.Equal(got, want)
}
