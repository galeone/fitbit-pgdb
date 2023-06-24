// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fitbit_pgdb

// package fitbit_pgdb contains an implementation of the [fitbit.Storage][1] interface
// for the [Fitbit Web API - Go client][2]
//
// [1]: https://github.com/galeone/fitbit/blob/main/storage.go
// [2]: https://github.com/galeone/fitbit

import (
	"github.com/galeone/fitbit/types"
	"github.com/galeone/igor"

	_ "github.com/joho/godotenv/autoload"
)

// PGDB implements the fitbit.Storage interface
type PGDB struct {
	*igor.Database
}

// NewPGDB creates a new connection to a PostgreSQL server.
// You must provide the connection string to use for connecting to a
// running PostgreSQL instance.
//
// It implements the `fitbit.Storage` interface.
func NewPGDB(connectionString string) *PGDB {
	var err error
	var db *igor.Database
	if db, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	return &PGDB{
		db,
	}
}

// InsertAuhorizingUser executes a CREATE query on the PostgreSQL database
// inserting the authorizing user in the associated table.
func (s *PGDB) InsertAuhorizingUser(authorizing *types.AuthorizingUser) error {
	authorizingUser := &AuthorizingUser{
		AuthorizingUser: *authorizing,
	}
	return s.Create(authorizingUser)
}

// UpsertAuthorizedUser executes INSERT or UPDATE the user on the PostgreSQL database.
// It inserts the user if the SELECT by user UserID fields returns the empty set.
// If it finds the user insteads it updates all the other NON PRIMARY KEY (or equivalent like UserID) fields.
func (s *PGDB) UpsertAuthorizedUser(authorized *types.AuthorizedUser) error {
	user := &AuthorizedUser{
		AuthorizedUser: *authorized,
	}
	var exists AuthorizedUser
	var err error
	var condition AuthorizedUser
	condition.UserID = authorized.UserID
	if err = s.Model(AuthorizedUser{}).Where(condition).Scan(&exists); err != nil {
		// First time we see this user
		err = s.Create(user)
	} else {
		user.ID = exists.ID
		err = s.Updates(user)
	}
	return err
}

// AuthorizedUser executes a SELECT for the types.AuthorizedUser using the
// `accessToken` as only query parameter.
// NOTE: the `accessToken` is unique on the underlying table.
func (s *PGDB) AuthorizedUser(accessToken string) (*types.AuthorizedUser, error) {
	var dbToken AuthorizedUser
	var err error
	var condition AuthorizedUser
	condition.AccessToken = accessToken
	if err = s.Model(AuthorizedUser{}).Where(&condition).Scan(&dbToken); err != nil {
		return nil, err
	}
	return &dbToken.AuthorizedUser, nil
}

// AuthorizingUser executes a SELECT for the types.AuthorizingUser using the `id`
// provided as PRIMARY KEY for creating the query.
func (s *PGDB) AuthorizingUser(id string) (*types.AuthorizingUser, error) {
	var authorizing AuthorizingUser
	var err error
	var condition AuthorizingUser
	condition.CSRFToken = id
	if err = s.Model(AuthorizingUser{}).Where(&condition).Scan(&authorizing); err != nil {
		return nil, err
	}
	return &types.AuthorizingUser{
		Code:      authorizing.Code,
		CSRFToken: authorizing.CSRFToken,
	}, nil
}
