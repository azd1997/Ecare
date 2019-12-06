package account

import "errors"

var (

	// UserId相关
	ErrInvalidUserId = errors.New("invalid UserId")

	// Role 相关
	ErrUnKnownQueryPattern = errors.New("unknown query pattern")
	ErrUnKnownRole = errors.New("unknown role")
	ErrNotRoleA = errors.New("not A type of role")
	ErrNotRoleB = errors.New("not B type of role")
	ErrUnmatchedRole = errors.New("unmatched role")
)
