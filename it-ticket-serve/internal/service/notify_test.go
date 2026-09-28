package service

import (
	"testing"

	"it-ticket-api/internal/model"
)

func TestTicketSubject(t *testing.T) {
	got := ticketSubject(&model.Ticket{ID: 5, Title: "腾讯地图无法使用"})
	if got != "#5 腾讯地图无法使用" {
		t.Fatalf("got %q", got)
	}
}

func TestWithReason(t *testing.T) {
	if got := withReason("剧芳 转派给你", ""); got != "剧芳 转派给你" {
		t.Fatalf("empty reason: %q", got)
	}
	if got := withReason("剧芳 已重开", "换完电源还是不亮"); got != "剧芳 已重开：换完电源还是不亮" {
		t.Fatalf("with reason: %q", got)
	}
}

func TestRoleCN(t *testing.T) {
	if roleCN("agent") != "IT" {
		t.Fatalf("agent: %q", roleCN("agent"))
	}
}
