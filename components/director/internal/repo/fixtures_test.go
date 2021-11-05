package repo_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	userTableName    = "users"
	appTableName     = "apps"
	bundlesTableName = "bundles"

	appID          = "appID"
	appName        = "appName"
	appDescription = "appDesc"

	bundleID          = "bundleID"
	bundleName        = "bundleName"
	bundleDescription = "bundleDesc"

	userID    = "given_id"
	tenantID  = "given_tenant"
	firstName = "given_first_name"
	lastName  = "given_last_name"
	age       = 55
)

var fixUser = User{
	ID:        userID,
	Tenant:    tenantID,
	FirstName: firstName,
	LastName:  lastName,
	Age:       age,
}

var fixApp = &App{
	ID:          appID,
	Name:        appName,
	Description: appDescription,
}

var fixBundle = &Bundle{
	ID:          bundleID,
	Name:        bundleName,
	Description: bundleDescription,
	AppID:       appID,
}

// User is a exemplary type to test generic Repositories
type User struct {
	ID        string `db:"id_col"`
	Tenant    string `db:"tenant_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}

func (a User) GetID() string {
	return a.ID
}

var userColumns = []string{"id_col", "tenant_id", "first_name", "last_name", "age"}

const UserType = resource.Type("UserType")

type UserCollection []User

func (u UserCollection) Len() int {
	return len(u)
}

type App struct {
	ID          string `db:"id_col"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (a *App) GetID() string {
	return a.ID
}

var appColumns = []string{"id_col", "name", "description"}

type AppCollection []App

func (a AppCollection) Len() int {
	return len(a)
}

type Bundle struct {
	ID          string `db:"id_col"`
	Name        string `db:"name"`
	Description string `db:"description"`
	AppID       string `db:"app_id"`
}

func (a *Bundle) GetID() string {
	return a.ID
}

func (a *Bundle) GetParentID() string {
	return a.AppID
}

var bundleColumns = []string{"id_col", "name", "description", "app_id"}

type BundleCollection []Bundle

func (a BundleCollection) Len() int {
	return len(a)
}

func someError() error {
	return errors.New("some error")
}
