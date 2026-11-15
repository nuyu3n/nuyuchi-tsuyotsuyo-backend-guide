package authz

type Role string

const (
	RoleLeader Role = "leader" // 団体の代表
	RoleStaff  Role = "staff"  // 運営
	RoleViewer Role = "viewer" // 閲覧のみの運営
)

type Action string

const (
	ActionView    Action = "application.view"
	ActionEdit    Action = "application.edit"
	ActionApprove Action = "application.approve"
)

// Assignment は「この行事で、この役割」を表す。
// OrgID は団体の代表のときだけ使う。
type Assignment struct {
	EventID int64
	Role    Role
	OrgID   int64
}

type Actor struct {
	UserID      int64
	Assignments []Assignment
}

type Application struct {
	ID      int64
	EventID int64
	OrgID   int64
}

var allowed = map[Role][]Action{
	RoleLeader: {ActionView, ActionEdit},
	RoleStaff:  {ActionView, ActionEdit, ActionApprove},
	RoleViewer: {ActionView},
}

func Can(a Actor, act Action, app Application) bool {
	for _, as := range a.Assignments {
		if as.EventID != app.EventID {
			continue
		}
		if as.Role == RoleLeader && as.OrgID != app.OrgID {
			continue
		}
		for _, x := range allowed[as.Role] {
			if x == act {
				return true
			}
		}
	}
	return false
}
