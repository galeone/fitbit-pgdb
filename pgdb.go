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
	"fmt"
	"os"

	"github.com/galeone/fitbit/types"
	"github.com/galeone/igor"

	_ "github.com/joho/godotenv/autoload"
)

// PGDB implements the fitbit.Storage interface
type PGDB struct {
	*igor.Database
}

// NewPGDB creates a new connection to a PostgreSQL server
// using the following environement variables:
// - DB_USER
// - DB_PASS
// - DB_NAME
// Loaded from a `.env` file, if present.
//
// It implements the `fitbit.Storage` interface.
func NewPGDB() *PGDB {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
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
	return s.Create(authorizing)
}

// UpsertAuthorizedUser executes INSERT or UPDATE the user on the PostgreSQL database.
// It insert the user, if the SELECT by user PRIMARY KEY returns the empty set.
// If it finds the user, it updates all the other NON PRIMARY KEY fields.
func (s *PGDB) UpsertAuthorizedUser(user *types.AuthorizedUser) error {
	var exists types.AuthorizedUser
	var err error
	if err = s.First(&exists, user.UserID); err != nil {
		// First time we see this user
		err = s.Create(user)
	} else {
		err = s.Updates(user)
	}
	return err
}

// AuthorizedUser executes a SELECT for the types.AuthorizedUser using the
// `accessToken` as only query parameter.
// NOTE: the `accessToken` is unique on the underlying table.
func (s *PGDB) AuthorizedUser(accessToken string) (*types.AuthorizedUser, error) {
	var dbToken types.AuthorizedUser
	var err error
	if err = s.Model(types.AuthorizedUser{}).Where(&types.AuthorizedUser{AccessToken: accessToken}).Scan(&dbToken); err != nil {
		return nil, err
	}
	return &dbToken, nil
}

// AuthorizingUser executes a SELECT for the types.AuthorizingUser using the `id`
// provided as PRIMARY KEY for creating the query.
func (s *PGDB) AuthorizingUser(id string) (*types.AuthorizingUser, error) {
	var authorizing types.AuthorizingUser
	var err error
	if err = s.First(&authorizing, id); err != nil {
		return nil, err
	}
	return &authorizing, nil
}
