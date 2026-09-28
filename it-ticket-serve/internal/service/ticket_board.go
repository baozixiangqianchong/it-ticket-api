package service

import (
	"time"

	"it-ticket-api/internal/logger"
	"it-ticket-api/internal/model"
	"it-ticket-api/internal/store"
)

func (s *TicketService) Board(role string) (*model.TicketBoard, error) {
	if role != "admin" {
		return nil, permissionDenied("只有管理员可以看工作台")
	}

	now := time.Now()
	admin := model.TicketFilter{ViewerRole: "admin"}

	pool, err := s.tickets.Count(model.TicketFilter{ViewerRole: "admin", Scope: model.ScopePool})
	if err != nil {
		return nil, err
	}
	inProgress, err := s.tickets.Count(withStatus(admin, model.StatusInProgress))
	if err != nil {
		return nil, err
	}
	pending, err := s.tickets.Count(withStatus(admin, model.StatusPending))
	if err != nil {
		return nil, err
	}
	resolved, err := s.tickets.Count(withStatus(admin, model.StatusResolved))
	if err != nil {
		return nil, err
	}
	poolStale, err := s.tickets.CountStaleOpen(now.Add(-poolStaleAfter))
	if err != nil {
		return nil, err
	}
	resolvedStale, err := s.tickets.CountStaleResolved(now.Add(-resolvedStaleAfter))
	if err != nil {
		return nil, err
	}
	agents, err := s.tickets.AgentLoads()
	if err != nil {
		return nil, err
	}
	if agents == nil {
		agents = []model.AgentLoad{}
	}

	return &model.TicketBoard{
		Pool:          pool,
		PoolStale:     poolStale,
		InProgress:    inProgress,
		Pending:       pending,
		Resolved:      resolved,
		ResolvedStale: resolvedStale,
		Agents:        agents,
	}, nil
}

func withStatus(f model.TicketFilter, status string) model.TicketFilter {
	f.Status = status
	return f
}

func (s *TicketService) CloseStaleResolved(actorID int64, role string) (*model.CloseStaleResult, error) {
	if role != "admin" {
		return nil, permissionDenied("只有管理员可以批量代关")
	}

	list, err := s.tickets.ListStaleResolved(time.Now().Add(-resolvedStaleAfter))
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return &model.CloseStaleResult{Closed: 0}, nil
	}

	tx, err := s.tickets.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	name := displayName(s.users, actorID)
	reason := closeStaleReason
	var closed int64
	for i := range list {
		t := &list[i]
		if err := s.tickets.ApplyTransition(tx, t.ID, store.TicketUpdate{
			Status:       model.StatusClosed,
			SetClosedNow: true,
		}); err != nil {
			return nil, err
		}
		from := t.Status
		if err := s.audits.Insert(tx, t.ID, actorID, model.ActionClose, &from, &reason, model.StatusClosed); err != nil {
			return nil, err
		}
		if err := s.notifyTicket(tx, t.CreatorID, actorID, t, model.NotifyClose, name+" 已关闭工单："+reason); err != nil {
			return nil, err
		}
		if t.AssigneeID != nil {
			if err := s.notifyTicket(tx, *t.AssigneeID, actorID, t, model.NotifyClose, name+" 已关闭工单："+reason); err != nil {
				return nil, err
			}
		}
		closed++
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	logger.Info("批量代关超期待确认", "actor_id", actorID, "closed", closed)
	return &model.CloseStaleResult{Closed: closed}, nil
}
