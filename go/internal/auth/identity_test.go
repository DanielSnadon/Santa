package auth

import "testing"

func TestIdentityV3(t *testing.T) {
	h, e := HashIdentityV3("correct horse battery staple")
	if e != nil || !VerifyIdentityV3(h, "correct horse battery staple") || VerifyIdentityV3(h, "wrong") {
		t.Fatal("identity hash verification failed")
	}
}
func TestNormalizeEmail(t *testing.T) {
	if NormalizeEmail(" A@Example.com ") != "A@EXAMPLE.COM" {
		t.Fatal("normalization")
	}
}
