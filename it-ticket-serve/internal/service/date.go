package service

import (
	"strings"
	"time"
)

// parseDayRange 把 YYYY-MM-DD 收成半开区间 [from, to)。to 为空表示不封顶。
func parseDayRange(fromDay, toDay string) (*time.Time, *time.Time, error) {
	fromDay = strings.TrimSpace(fromDay)
	toDay = strings.TrimSpace(toDay)
	var from, to *time.Time
	if fromDay != "" {
		t, err := time.ParseInLocation("2006-01-02", fromDay, time.Local)
		if err != nil {
			return nil, nil, invalidArg("开始日期格式应为 YYYY-MM-DD")
		}
		from = &t
	}
	if toDay != "" {
		t, err := time.ParseInLocation("2006-01-02", toDay, time.Local)
		if err != nil {
			return nil, nil, invalidArg("结束日期格式应为 YYYY-MM-DD")
		}
		end := t.Add(24 * time.Hour)
		to = &end
	}
	if from != nil && to != nil && !from.Before(*to) {
		return nil, nil, invalidArg("开始日期不能晚于结束日期")
	}
	return from, to, nil
}

func validTimeField(s string) bool {
	switch s {
	case "", "created_at", "updated_at", "closed_at":
		return true
	default:
		return false
	}
}
