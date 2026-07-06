package domain

type UserPermission string

const (
	UserPermissionDeleteChat UserPermission = "delete chat"
	UserPermissionAddUser    UserPermission = "add user"
	UserPermissionDeleteUser UserPermission = "delete user"
)

func (u UserPermission) IsValid() bool {
	switch u {
	case UserPermissionDeleteChat, UserPermissionAddUser, UserPermissionDeleteUser:
		return true
	}

	return false
}
