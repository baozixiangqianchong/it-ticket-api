package service

import (
	"crypto/rand"
	"time"

	"it-ticket-api/internal/model"
)

const inviteTTL = 24 * time.Hour

func normalizeInviteInput(s string) string {
	return model.NormalizeInviteCode(s)
}

func randomInviteCode() (string, error) {
	const alphabet = model.InviteAlphabet
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	out := make([]byte, 8)
	for i, b := range raw {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out), nil
}

func (s *AdminService) CreateInvite(actorID int64) (*model.PublicInvite, error) {
	var last error
	for i := 0; i < 5; i++ {
		code, err := randomInviteCode()
		if err != nil {
			return nil, err
		}
		inv, err := s.invites.Insert(actorID, code, time.Now().Add(inviteTTL))
		if err != nil {
			last = err
			continue
		}
		pub := inv.Public(time.Now())
		return &pub, nil
	}
	return nil, last
}

func (s *AdminService) ListInvites() (*model.InviteList, error) {
	list, err := s.invites.List(30)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]model.PublicInvite, 0, len(list))
	for _, inv := range list {
		items = append(items, inv.Public(now))
	}
	return &model.InviteList{Items: items}, nil
}
