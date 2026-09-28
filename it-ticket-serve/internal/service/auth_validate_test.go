package service

import "testing"

func TestNormalizeDisplayName(t *testing.T) {
	got, err := normalizeDisplayName("  花名  ")
	if err != nil || got != "花名" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := normalizeDisplayName("   "); err == nil {
		t.Fatal("空显示名应失败")
	}
}

func TestValidatePasswordLength(t *testing.T) {
	if err := validatePasswordLength("short"); err == nil {
		t.Fatal("过短密码应失败")
	}
	if err := validatePasswordLength("password1"); err != nil {
		t.Fatal(err)
	}
}
