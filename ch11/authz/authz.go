package authz

// Role は、string に名前をつけた型。ただの文字列と取りちがえにくくなる。
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

// 役割ごとに、許す操作を並べる。ここに書いていない操作はすべて禁止。
var allowed = map[Role][]Action{
	RoleLeader: {ActionView, ActionEdit},
	RoleStaff:  {ActionView, ActionEdit, ActionApprove},
	RoleViewer: {ActionView},
}

// Can は、a が app に対して act をしてよいかを返す。
func Can(a Actor, act Action, app Application) bool {
	for _, as := range a.Assignments {
		// 別の行事の役割は関係ない
		if as.EventID != app.EventID {
			continue
		}
		// 代表が触れるのは、自分の団体の申し込みだけ
		if as.Role == RoleLeader && as.OrgID != app.OrgID {
			continue
		}
		// この役割で許されている操作か調べる
		for _, x := range allowed[as.Role] {
			if x == act {
				return true
			}
		}
	}
	// どの役割でも許されなければ禁止
	return false
}
