package service

import (
	"testing"
	"time"
)

func TestParseDayRange_SameDay(t *testing.T) {
	from, to, err := parseDayRange("2026-09-21", "2026-09-21")
	if err != nil {
		t.Fatal(err)
	}
	if from == nil || to == nil {
		t.Fatal("起止都应该有值")
	}
	if to.Sub(*from) != 24*time.Hour {
		t.Fatalf("同一天应是 24 小时半开区间, got %s", to.Sub(*from))
	}
}

func TestParseDayRange_Order(t *testing.T) {
	if _, _, err := parseDayRange("2026-09-28", "2026-09-21"); err == nil {
		t.Fatal("开始晚于结束应失败")
	}
}
