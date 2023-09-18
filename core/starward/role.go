package starward

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleNormal  Role = "normal"
	RoleGuest   Role = "guest"
)

func (r Role) ExportDatabase() string {
	return string(r)
}
