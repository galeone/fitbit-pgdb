// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fitbit_pgdb_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	pgdb "github.com/galeone/fitbit-pgdb/v3"
	"github.com/galeone/fitbit/v2"
	"github.com/galeone/fitbit/v2/types"
	"github.com/galeone/igor"
)

var _connectionString string

func init() {

	var err error

	_connectionString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	var igordb *igor.Database
	if igordb, err = igor.Connect(_connectionString); err != nil {
		panic(fmt.Sprintf("%s: %s", _connectionString, err.Error()))
	}

	tx := igordb.Begin()

	var authorizedUser pgdb.AuthorizedUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		id BIGSERIAL PRIMARY KEY,
		user_id TEXT NOT NULL,
		token_type TEXT NOT NULL,
		scope TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_in INTEGER NOT NULL,
		access_token TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(access_token),
		UNIQUE(user_id)
	)`, authorizedUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	var authorizingUser pgdb.AuthorizingUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		id BIGSERIAL PRIMARY KEY,
		csrftoken TEXT NOT NULL,
		code TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(csrftoken)
	)`, authorizingUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Commit(); err != nil {
		panic(err.Error())
	}
}

var db fitbit.Storage

func TestPGDBNew(t *testing.T) {
	// Test connection re-use
	if connection, err := sql.Open("postgres", _connectionString); err != nil {
		if db = pgdb.NewPGDBFromConnection(connection); db == nil {
			t.Errorf("Unable to re-use db connection when creating PGDB object")
		}
	}

	// Test connection creation
	if db = pgdb.NewPGDB(_connectionString); db == nil {
		t.Errorf("Unable to create a PGDB object creating a connection")
	}

}

func TestAuthorizingUser(t *testing.T) {
	db = pgdb.NewPGDB(_connectionString)
	var err error
	pk := "unique token"
	if err = db.InsertAuthorizingUser(&types.AuthorizingUser{
		Code:      "1",
		CSRFToken: pk,
	}); err != nil {
		t.Errorf("InsertAuthorizingUser should succeed but got: %s", err)
	}
	var user *types.AuthorizingUser
	if user, err = db.AuthorizingUser(pk); err != nil {
		t.Errorf("AuthorizingUser (get by ID) should work but got: %s", err)
	}
	if user.Code != "1" {
		t.Errorf("Expected a correct retrieval of the authorizing user, but got this user instead %v", user)
	}
}

func TestAuthorized(t *testing.T) {
	db = pgdb.NewPGDB(_connectionString)
	var err error
	pk := "unique id"
	user := types.AuthorizedUser{
		UserID:       "random",
		AccessToken:  pk,
		RefreshToken: "something else",
		ExpiresIn:    100,
		Scope:        "list of scopes",
		TokenType:    "Bearer",
	}

	// Insert
	if err = db.UpsertAuthorizedUser(&user); err != nil {
		t.Errorf("Insert of Auhtorized User should work but got: %s", err)
	}

	// Fetch
	var fetched *types.AuthorizedUser
	if fetched, err = db.AuthorizedUser(pk); err != nil {
		t.Errorf("AuthorizedUser (get by ID) should work but got: %s", err)
	}

	if fetched.AccessToken != user.AccessToken {
		t.Errorf("Fetched user differs from inserted user. Expected %v got %v", user, fetched)
	}

	// Update
	user.RefreshToken = "changed"
	if err = db.UpsertAuthorizedUser(&user); err != nil {
		t.Errorf("Update of Auhtorized User should work but got: %s", err)
	}

	// Verify
	if fetched, err = db.AuthorizedUser(pk); err != nil {
		t.Errorf("AuthorizedUser (get by ID) should work but got: %s", err)
	}

	if fetched.RefreshToken != user.RefreshToken {
		t.Errorf("Fetched user after update differs from inserted user. Expected %v got %v", user, fetched)
	}
}
