package fitbit_pgdb

import (
	"time"

	"github.com/galeone/fitbit/types"
)

// AuthorizedUser wraps fitbit types.AuthorizedUser
// and adds:
// - The igor-required method TableName()
// - The primary key
// - The created at field
type AuthorizedUser struct {
	types.AuthorizedUser
	ID        int64     `igor:"primary_key"`
	CreatedAt time.Time `sql:"default:now()"`
}

func (AuthorizedUser) TableName() string {
	return "oauth2_authorized"
}

// AuthorizingUser wraps fitbit types.AuthorizingUser
// and adds:
// - The igor-required method TableName()
// - The primary key
// - The created at field
type AuthorizingUser struct {
	types.AuthorizingUser
	ID        int64     `igor:"primary_key"`
	CreatedAt time.Time `sql:"default:now()"`
}

func (AuthorizingUser) TableName() string {
	return "oauth2_authorizing"
}
