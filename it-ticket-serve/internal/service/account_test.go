package service

import "testing"

func TestCanHandleTickets(t *testing.T) {
	if !canHandleTickets("agent") || !canHandleTickets("admin") {
		t.Fatal("IT 和管理员都应能接单")
	}
	if canHandleTickets("user") {
		t.Fatal("员工不能接单")
	}
}

func TestShouldReleaseTickets(t *testing.T) {
	if shouldReleaseTickets("agent", "admin", false) {
		t.Fatal("升管理员应留下手上的单")
	}
	if shouldReleaseTickets("admin", "agent", false) {
		t.Fatal("管理员改回 IT 应留下手上的单")
	}
	if !shouldReleaseTickets("agent", "user", false) {
		t.Fatal("取消 IT 应退回未关单")
	}
	if !shouldReleaseTickets("admin", "user", false) {
		t.Fatal("管理员改成员工应退回未关单")
	}
	if !shouldReleaseTickets("agent", "agent", true) {
		t.Fatal("停用处理人应退回未关单")
	}
	if shouldReleaseTickets("user", "user", true) {
		t.Fatal("停用普通员工没有要退的单")
	}
}

func TestNormalizeInviteCode(t *testing.T) {
	if got := normalizeInviteInput("ab2d-ef3g"); got != "AB2DEF3G" {
		t.Fatalf("got %q", got)
	}
}
