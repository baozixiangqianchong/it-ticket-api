package store

import (
	"strings"
	"testing"
	"time"

	"it-ticket-api/internal/model"
)

func TestFilterSQL_SearchCoversDescriptionAndComments(t *testing.T) {
	where, args := filterSQL(model.TicketFilter{ViewerRole: "admin", Q: "打印机卡纸"})
	if !strings.Contains(where, "t.title LIKE") || !strings.Contains(where, "t.description LIKE") {
		t.Fatalf("搜索应覆盖标题和描述: %s", where)
	}
	if !strings.Contains(where, "ticket_comments") {
		t.Fatalf("搜索应覆盖评论: %s", where)
	}
	if len(args) != 3 {
		t.Fatalf("期望 3 个 like 参数，得到 %d (%#v)", len(args), args)
	}
}

func TestFilterSQL_TimeRangeUsesClosedAt(t *testing.T) {
	from := time.Date(2026, 9, 21, 0, 0, 0, 0, time.Local)
	to := from.Add(7 * 24 * time.Hour)
	where, args := filterSQL(model.TicketFilter{
		ViewerRole: "admin",
		TimeField:  "closed_at",
		From:       &from,
		To:         &to,
	})
	if !strings.Contains(where, "t.closed_at >= ?") || !strings.Contains(where, "t.closed_at < ?") {
		t.Fatalf("关闭时间筛选不对: %s", where)
	}
	if len(args) != 2 {
		t.Fatalf("时间参数数量不对: %#v", args)
	}
}

func TestFilterSQL_NumericSearchIncludesID(t *testing.T) {
	where, args := filterSQL(model.TicketFilter{ViewerRole: "admin", Q: "#12"})
	if !strings.Contains(where, "t.id = ?") {
		t.Fatalf("数字搜索应带单号: %s", where)
	}
	if len(args) == 0 || args[0] != int64(12) {
		t.Fatalf("单号参数不对: %#v", args)
	}
}
