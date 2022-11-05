// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package fitbit_pgdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/galeone/fitbit"
	pgdb "github.com/galeone/fitbit-pgdb"
	"github.com/galeone/fitbit/types"
	"github.com/galeone/igor"
)

func init() {

	var err error

	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))
	var igordb *igor.Database
	if igordb, err = igor.Connect(connectionString); err != nil {
		panic(err.Error())
	}

	tx := igordb.Begin()

	var authorizedUser types.AuthorizedUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		user_id TEXT NOT NULL PRIMARY KEY,
		token_type TEXT NOT NULL,
		scope TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		expires_in INTEGER NOT NULL,
		access_token TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		UNIQUE(access_token))`, authorizedUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	var authorizingUser types.AuthorizingUser
	if err = tx.Exec(fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS "%s" (
		csrftoken TEXT NOT NULL PRIMARY KEY,
		code TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`, authorizingUser.TableName())); err != nil {
		_ = tx.Rollback()
		panic(err.Error())
	}

	if err = tx.Commit(); err != nil {
		panic(err.Error())
	}
}

var db fitbit.Storage

func TestAuthorizingUser(t *testing.T) {
	db = pgdb.NewPGDB()
	var err error
	pk := "primary key value"
	if err = db.InsertAuhorizingUser(&types.AuthorizingUser{
		Code:      "1",
		CSRFToken: pk,
	}); err != nil {
		t.Errorf("InsertAuhorizingUser should succeed but got: %s", err)
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
	db = pgdb.NewPGDB()
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
