package service

import (
	"strings"
	"time"
	"unicode/utf8"

	"it-ticket-api/internal/model"
)

// reopenWindow 关闭后还能重开的宽限期。超时只能看，不能再改。
const reopenWindow = 7 * 24 * time.Hour
const poolStaleAfter = 2 * 24 * time.Hour
const resolvedStaleAfter = 7 * 24 * time.Hour
const closeStaleReason = "超过 7 天未确认，管理员批量代关"

// actionResult 状态机算完的下一步。写库由 Update 按这些标志拼 SQL。
type actionResult struct {
	To            string
	ClearAssignee bool
	SetClosedAt   bool
	ClearClosedAt bool
}

// resolveTicketAction 判断这一步能不能走、走到哪。非法跳转 409，人不对 403。
// now 用来算关单后的 7 天窗口，测试里可以钉死。
func resolveTicketAction(t *model.Ticket, actorID int64, role, action, reason string, now time.Time) (actionResult, error) {
	isAdmin := role == "admin"
	isAssignee := t.AssigneeID != nil && *t.AssigneeID == actorID
	isCreator := t.CreatorID == actorID
	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) > 500 {
		return actionResult{}, invalidArg("原因最多 500 个字")
	}

	switch action {
	case model.ActionStart:
		if t.Status != model.StatusAssigned {
			return actionResult{}, ticketInvalidTransition("当前状态不能开始处理")
		}
		if !isAdmin && !isAssignee {
			return actionResult{}, permissionDenied("只有当前处理人或管理员可以开始处理")
		}
		return actionResult{To: model.StatusInProgress}, nil

	case model.ActionResolve:
		if t.Status != model.StatusInProgress {
			return actionResult{}, ticketInvalidTransition("当前状态不能标记已解决")
		}
		if !isAdmin && !isAssignee {
			return actionResult{}, permissionDenied("只有当前处理人或管理员可以标记已解决")
		}
		return actionResult{To: model.StatusResolved}, nil

	case model.ActionClose:
		if t.Status != model.StatusResolved {
			return actionResult{}, ticketInvalidTransition("当前状态不能关闭")
		}
		if !isAdmin && !isCreator {
			return actionResult{}, permissionDenied("只有提单人或管理员可以关闭工单")
		}
		if isAdmin && !isCreator && reason == "" {
			return actionResult{}, invalidArg("管理员代关必须填写原因")
		}
		return actionResult{To: model.StatusClosed, SetClosedAt: true}, nil

	case model.ActionReopen:
		if t.Status != model.StatusResolved && t.Status != model.StatusClosed {
			return actionResult{}, ticketInvalidTransition("当前状态不能重开")
		}
		if t.Status == model.StatusClosed && !withinReopenWindow(t.ClosedAt, now) {
			return actionResult{}, ticketClosed("关闭已超过 7 天，不能再重开")
		}
		if !isAdmin && !isCreator {
			return actionResult{}, permissionDenied("只有提单人或管理员可以重开工单")
		}
		if reason == "" {
			return actionResult{}, invalidArg("重开必须填写原因")
		}
		// 默认留下处理人，回到 assigned。处理人失效时由 Update 改成 open。
		return actionResult{To: model.StatusAssigned, ClearClosedAt: true}, nil

	case model.ActionCancel:
		switch t.Status {
		case model.StatusOpen, model.StatusAssigned, model.StatusInProgress:
		default:
			return actionResult{}, ticketInvalidTransition("等用户或已解决后不能撤回，请走关闭或重开")
		}
		if !isCreator {
			return actionResult{}, permissionDenied("只有提单人可以撤回工单")
		}
		return actionResult{To: model.StatusClosed, SetClosedAt: true}, nil

	case model.ActionWait:
		if t.Status != model.StatusInProgress {
			return actionResult{}, ticketInvalidTransition("当前状态不能标记等用户")
		}
		if !isAdmin && !isAssignee {
			return actionResult{}, permissionDenied("只有当前处理人或管理员可以标记等用户")
		}
		if reason == "" {
			return actionResult{}, invalidArg("等用户必须写明需要补充什么")
		}
		return actionResult{To: model.StatusPending}, nil

	case model.ActionResume:
		if t.Status != model.StatusPending {
			return actionResult{}, ticketInvalidTransition("当前状态不能继续处理")
		}
		if !isAdmin && !isAssignee {
			return actionResult{}, permissionDenied("只有当前处理人或管理员可以继续处理")
		}
		return actionResult{To: model.StatusInProgress}, nil

	default:
		return actionResult{}, invalidArg("动作必须是 start / resolve / close / reopen / cancel / wait / resume")
	}
}

func withinReopenWindow(closedAt *time.Time, now time.Time) bool {
	if closedAt == nil {
		return false
	}
	return !now.After(closedAt.Add(reopenWindow))
}

// availableActions 当前用户此刻能点的按钮。详情和列表共用，页面不要再复制状态机。
func availableActions(t *model.Ticket, actorID int64, role string, now time.Time) []string {
	isAdmin := role == "admin"
	isAssignee := t.AssigneeID != nil && *t.AssigneeID == actorID
	isCreator := t.CreatorID == actorID
	out := make([]string, 0, 4)

	canAssign := t.Status == model.StatusOpen || t.Status == model.StatusAssigned || t.Status == model.StatusInProgress || t.Status == model.StatusPending
	if isAdmin && canAssign {
		out = append(out, model.ActionAssign)
	}
	if (role == "agent" || isAdmin) && t.Status == model.StatusOpen && t.AssigneeID == nil {
		out = append(out, model.ActionClaim)
	}
	if t.Status == model.StatusAssigned && (isAdmin || isAssignee) {
		out = append(out, model.ActionStart)
	}
	if t.Status == model.StatusInProgress && (isAdmin || isAssignee) {
		out = append(out, model.ActionResolve, model.ActionWait)
	}
	if t.Status == model.StatusPending && (isAdmin || isAssignee) {
		out = append(out, model.ActionResume)
	}
	if (isAdmin || isAssignee) && (t.Status == model.StatusAssigned || t.Status == model.StatusInProgress || t.Status == model.StatusPending) {
		out = append(out, model.ActionTransfer)
	}
	if t.Status == model.StatusResolved && (isAdmin || isCreator) {
		out = append(out, model.ActionClose, model.ActionReopen)
	}
	if t.Status == model.StatusClosed && (isAdmin || isCreator) && withinReopenWindow(t.ClosedAt, now) {
		out = append(out, model.ActionReopen)
	}
	if isCreator && (t.Status == model.StatusOpen || t.Status == model.StatusAssigned || t.Status == model.StatusInProgress) {
		out = append(out, model.ActionCancel)
	}
	if t.Status != model.StatusClosed {
		if t.Status == model.StatusOpen && (isCreator || isAdmin) {
			out = append(out, model.ActionEdit)
		} else if isAdmin || isAssignee {
			out = append(out, model.ActionEdit)
		}
	}
	return out
}

func ticketStale(t *model.Ticket, now time.Time) bool {
	switch t.Status {
	case model.StatusOpen:
		return t.AssigneeID == nil && now.Sub(t.CreatedAt) > poolStaleAfter
	case model.StatusResolved:
		return now.Sub(t.UpdatedAt) > resolvedStaleAfter
	default:
		return false
	}
}

// statusAfterHandoff 领/派/转之后：等用户的单保持等待，其余直接进入处理中。
func statusAfterHandoff(from string) string {
	if from == model.StatusPending {
		return model.StatusPending
	}
	return model.StatusInProgress
}

func reasonPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func validTicketStatus(s string) bool {
	switch s {
	case model.StatusOpen, model.StatusAssigned, model.StatusInProgress, model.StatusPending, model.StatusResolved, model.StatusClosed:
		return true
	default:
		return false
	}
}
