package authz

import "testing"

func TestCan(t *testing.T) {
	const event, otherEvent = 1, 2
	const myOrg, otherOrg = 10, 20

	leader := Actor{UserID: 100, Assignments: []Assignment{{EventID: event, Role: RoleLeader, OrgID: myOrg}}}
	staff := Actor{UserID: 200, Assignments: []Assignment{{EventID: event, Role: RoleStaff}}}
	viewer := Actor{UserID: 300, Assignments: []Assignment{{EventID: event, Role: RoleViewer}}}
	nobody := Actor{UserID: 400}

	mine := Application{ID: 1, EventID: event, OrgID: myOrg}
	others := Application{ID: 2, EventID: event, OrgID: otherOrg}
	otherEvents := Application{ID: 3, EventID: otherEvent, OrgID: myOrg}

	tests := []struct {
		name  string
		actor Actor
		act   Action
		app   Application
		want  bool
	}{
		{"代表は自分の団体を見られる", leader, ActionView, mine, true},
		{"代表は自分の団体を編集できる", leader, ActionEdit, mine, true},
		{"代表は他の団体を見られない", leader, ActionView, others, false},
		{"代表は承認できない", leader, ActionApprove, mine, false},
		{"代表は別の行事を見られない", leader, ActionView, otherEvents, false},
		{"運営は他の団体を見られる", staff, ActionView, others, true},
		{"運営は承認できる", staff, ActionApprove, others, true},
		{"運営は別の行事を見られない", staff, ActionView, otherEvents, false},
		{"閲覧のみは見られる", viewer, ActionView, others, true},
		{"閲覧のみは承認できない", viewer, ActionApprove, others, false},
		{"役割のない人は何もできない", nobody, ActionView, mine, false},
		{"知らない操作はできない", staff, Action("application.delete"), mine, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Can(tt.actor, tt.act, tt.app)
			if got != tt.want {
				t.Errorf("Can() = %v, want %v", got, tt.want)
			}
		})
	}
}
