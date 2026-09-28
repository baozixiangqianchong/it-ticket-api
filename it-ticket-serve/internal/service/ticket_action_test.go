package service

import (
	"testing"
	"time"

	"it-ticket-api/internal/model"
	"it-ticket-api/internal/response"
)

func TestResolveTicketAction(t *testing.T) {
	assignee := int64(8)
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	closedOK := now.Add(-3 * 24 * time.Hour)
	closedLate := now.Add(-8 * 24 * time.Hour)

	tests := []struct {
		name    string
		ticket  model.Ticket
		actor   int64
		role    string
		action  string
		reason  string
		wantTo  string
		wantErr string
	}{
		{
			name:   "处理人开始修",
			ticket: model.Ticket{Status: model.StatusAssigned, AssigneeID: &assignee, CreatorID: 3},
			actor:  8, role: "agent", action: model.ActionStart,
			wantTo: model.StatusInProgress,
		},
		{
			name:   "open 不能直接 resolve",
			ticket: model.Ticket{Status: model.StatusOpen, CreatorID: 3},
			actor:  8, role: "agent", action: model.ActionResolve,
			wantErr: response.ErrTicketInvalidTrans,
		},
		{
			name:   "员工关闭自己的单不必写原因",
			ticket: model.Ticket{Status: model.StatusResolved, CreatorID: 3, AssigneeID: &assignee},
			actor:  3, role: "user", action: model.ActionClose,
			wantTo: model.StatusClosed,
		},
		{
			name:   "管理员代关必须写原因",
			ticket: model.Ticket{Status: model.StatusResolved, CreatorID: 3, AssigneeID: &assignee},
			actor:  1, role: "admin", action: model.ActionClose,
			wantErr: response.ErrInvalidArgument,
		},
		{
			name:   "管理员代关写了原因",
			ticket: model.Ticket{Status: model.StatusResolved, CreatorID: 3, AssigneeID: &assignee},
			actor:  1, role: "admin", action: model.ActionClose, reason: "用户已离职",
			wantTo: model.StatusClosed,
		},
		{
			name:   "重开必须写原因",
			ticket: model.Ticket{Status: model.StatusResolved, CreatorID: 3, AssigneeID: &assignee},
			actor:  3, role: "user", action: model.ActionReopen,
			wantErr: response.ErrInvalidArgument,
		},
		{
			name:   "resolved 重开回到 assigned",
			ticket: model.Ticket{Status: model.StatusResolved, CreatorID: 3, AssigneeID: &assignee},
			actor:  3, role: "user", action: model.ActionReopen, reason: "还是不亮",
			wantTo: model.StatusAssigned,
		},
		{
			name:   "关单 3 天内可重开",
			ticket: model.Ticket{Status: model.StatusClosed, CreatorID: 3, AssigneeID: &assignee, ClosedAt: &closedOK},
			actor:  3, role: "user", action: model.ActionReopen, reason: "关早了",
			wantTo: model.StatusAssigned,
		},
		{
			name:   "关单超过 7 天不能重开",
			ticket: model.Ticket{Status: model.StatusClosed, CreatorID: 3, AssigneeID: &assignee, ClosedAt: &closedLate},
			actor:  3, role: "user", action: model.ActionReopen, reason: "还想重开",
			wantErr: response.ErrTicketClosed,
		},
		{
			name:   "创建人撤回 open 单",
			ticket: model.Ticket{Status: model.StatusOpen, CreatorID: 3},
			actor:  3, role: "user", action: model.ActionCancel,
			wantTo: model.StatusClosed,
		},
		{
			name:   "处理中创建人可撤回",
			ticket: model.Ticket{Status: model.StatusInProgress, CreatorID: 3, AssigneeID: &assignee},
			actor:  3, role: "user", action: model.ActionCancel,
			wantTo: model.StatusClosed,
		},
		{
			name:   "等用户后不能撤回",
			ticket: model.Ticket{Status: model.StatusPending, CreatorID: 3, AssigneeID: &assignee},
			actor:  3, role: "user", action: model.ActionCancel,
			wantErr: response.ErrTicketInvalidTrans,
		},
		{
			name:   "等用户必须写原因",
			ticket: model.Ticket{Status: model.StatusInProgress, CreatorID: 3, AssigneeID: &assignee},
			actor:  8, role: "agent", action: model.ActionWait,
			wantErr: response.ErrInvalidArgument,
		},
		{
			name:   "非处理人不能开始修",
			ticket: model.Ticket{Status: model.StatusAssigned, CreatorID: 3, AssigneeID: &assignee},
			actor:  9, role: "agent", action: model.ActionStart,
			wantErr: response.ErrPermissionDenied,
		},
		{
			name:   "处理人标记等用户",
			ticket: model.Ticket{Status: model.StatusInProgress, CreatorID: 3, AssigneeID: &assignee},
			actor:  8, role: "agent", action: model.ActionWait, reason: "请补充报错截图",
			wantTo: model.StatusPending,
		},
		{
			name:   "处理人从 pending 继续修",
			ticket: model.Ticket{Status: model.StatusPending, CreatorID: 3, AssigneeID: &assignee},
			actor:  8, role: "agent", action: model.ActionResume,
			wantTo: model.StatusInProgress,
		},
		{
			name:   "open 不能 wait",
			ticket: model.Ticket{Status: model.StatusOpen, CreatorID: 3},
			actor:  8, role: "agent", action: model.ActionWait,
			wantErr: response.ErrTicketInvalidTrans,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTicketAction(&tt.ticket, tt.actor, tt.role, tt.action, tt.reason, now)
			if tt.wantErr != "" {
				var apiErr *Error
				if err == nil || !asError(err, &apiErr) || apiErr.Name != tt.wantErr {
					t.Fatalf("want error %s, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.To != tt.wantTo {
				t.Fatalf("to = %s, want %s", got.To, tt.wantTo)
			}
		})
	}
}

func TestAvailableActions(t *testing.T) {
	assignee := int64(8)
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	open := model.Ticket{Status: model.StatusOpen, CreatorID: 3}

	got := availableActions(&open, 8, "agent", now)
	if !contains(got, model.ActionClaim) || contains(got, model.ActionCancel) {
		t.Fatalf("agent 在待派池应能 claim，不能 cancel: %v", got)
	}

	got = availableActions(&open, 3, "user", now)
	if !contains(got, model.ActionCancel) || contains(got, model.ActionClaim) {
		t.Fatalf("员工在自己的 open 单应能 cancel，不能 claim: %v", got)
	}

	assigned := model.Ticket{Status: model.StatusAssigned, CreatorID: 3, AssigneeID: &assignee}
	got = availableActions(&assigned, 1, "admin", now)
	if !contains(got, model.ActionAssign) || !contains(got, model.ActionStart) {
		t.Fatalf("管理员在 assigned 应能 assign 和 start: %v", got)
	}

	inProgress := model.Ticket{Status: model.StatusInProgress, CreatorID: 3, AssigneeID: &assignee}
	got = availableActions(&inProgress, 8, "agent", now)
	if !contains(got, model.ActionWait) || !contains(got, model.ActionTransfer) {
		t.Fatalf("处理人在 in_progress 应能 wait 和 transfer: %v", got)
	}

	got = availableActions(&inProgress, 3, "user", now)
	if !contains(got, model.ActionCancel) {
		t.Fatalf("提单人在处理中应能撤回: %v", got)
	}

	if statusAfterHandoff(model.StatusPending) != model.StatusPending {
		t.Fatalf("转派 pending 应保持等用户")
	}
	if statusAfterHandoff(model.StatusOpen) != model.StatusInProgress {
		t.Fatalf("领取 open 应进入处理中")
	}

	got = availableActions(&open, 3, "user", now)
	if !contains(got, model.ActionEdit) {
		t.Fatalf("提单人在 open 应能改单: %v", got)
	}
	got = availableActions(&inProgress, 3, "user", now)
	if contains(got, model.ActionEdit) {
		t.Fatalf("提单人在处理中不应改单: %v", got)
	}
}

func TestPrepareTicketEdit(t *testing.T) {
	open := model.Ticket{ID: 1, Title: "旧标题", Description: "旧描述", Category: model.CategoryNetwork, Priority: model.PriorityP2, Status: model.StatusOpen, CreatorID: 3}
	_, _, _, _, _, err := prepareTicketEdit(&open, 3, "user", model.EditTicketInput{
		Title: "新标题", Description: "新描述", Category: model.CategorySoftware, Priority: model.PriorityP1,
	})
	if err == nil {
		t.Fatal("员工不能把单标成紧急")
	}

	title, _, _, pri, summary, err := prepareTicketEdit(&open, 3, "user", model.EditTicketInput{
		Title: "新标题", Description: "新描述", Category: model.CategorySoftware, Priority: model.PriorityP2,
	})
	if err != nil || title != "新标题" || pri != model.PriorityP2 || summary == "" {
		t.Fatalf("open 改单失败: title=%s pri=%s summary=%s err=%v", title, pri, summary, err)
	}

	inProgress := open
	inProgress.Status = model.StatusInProgress
	assignee := int64(8)
	inProgress.AssigneeID = &assignee
	_, _, _, _, _, err = prepareTicketEdit(&inProgress, 3, "user", model.EditTicketInput{
		Title: "再改", Description: "新描述", Category: model.CategorySoftware, Priority: model.PriorityP3,
	})
	if err == nil {
		t.Fatal("处理中提单人不应能改")
	}

	title, desc, _, pri, _, err := prepareTicketEdit(&inProgress, 8, "agent", model.EditTicketInput{
		Title: "黑客标题", Description: "黑客描述", Category: model.CategoryHardware, Priority: model.PriorityP3,
	})
	if err != nil || title != open.Title || desc != open.Description || pri != model.PriorityP3 {
		t.Fatalf("处理人只能改优先级: title=%s desc=%s pri=%s err=%v", title, desc, pri, err)
	}
}

func asError(err error, dest **Error) bool {
	e, ok := err.(*Error)
	if !ok {
		return false
	}
	*dest = e
	return true
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
