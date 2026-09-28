package model

import (
	"strings"
	"time"
)

const InviteAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func NormalizeInviteCode(s string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(strings.TrimSpace(s)) {
		if r == '-' || r == ' ' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func FormatInviteCode(s string) string {
	s = NormalizeInviteCode(s)
	if len(s) != 8 {
		return s
	}
	return s[:4] + "-" + s[4:]
}

func ValidInviteCode(s string) bool {
	s = NormalizeInviteCode(s)
	if len(s) != 8 {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune(InviteAlphabet, r) {
			return false
		}
	}
	return true
}

const (
	InviteUnused  = "unused"
	InviteUsed    = "used"
	InviteExpired = "expired"
)

type InviteCode struct {
	ID          int64
	Code        string
	CreatedBy   int64
	CreatorName string
	ExpiresAt   time.Time
	UsedAt      *time.Time
	UsedBy      *int64
	UsedName    *string
	CreatedAt   time.Time
}

type PublicInvite struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsedAt      *time.Time `json:"used_at"`
	UsedByName  *string    `json:"used_by_name"`
	CreatorName string     `json:"creator_name"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
}

type InviteList struct {
	Items []PublicInvite `json:"items"`
}

func (i InviteCode) Status(now time.Time) string {
	if i.UsedAt != nil {
		return InviteUsed
	}
	if !i.ExpiresAt.After(now) {
		return InviteExpired
	}
	return InviteUnused
}

func (i InviteCode) Public(now time.Time) PublicInvite {
	return PublicInvite{
		ID:          i.ID,
		Code:        FormatInviteCode(i.Code),
		ExpiresAt:   i.ExpiresAt,
		UsedAt:      i.UsedAt,
		UsedByName:  i.UsedName,
		CreatorName: i.CreatorName,
		Status:      i.Status(now),
		CreatedAt:   i.CreatedAt,
	}
}
