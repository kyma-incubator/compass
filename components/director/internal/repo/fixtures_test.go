package repo_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	userTableName     = "users"
	appTableName      = "apps"
	bundlesTableName  = "bundles"
	webhooksTableName = "webhooks"
	biaTableName      = "bia"

	appID           = "appID"
	appID2          = "appID2"
	appName         = "appName"
	appName2        = "appName"
	appDescription  = "appDesc"
	appDescription2 = "appDesc"

	bundleID          = "bundleID"
	bundleName        = "bundleName"
	bundleDescription = "bundleDesc"

	biaID          = "biaID"
	biaName        = "biaName"
	biaDescription = "biaDesc"

	whID = "whID"
	ftID = "ftID"

	userID    = "given_id"
	tenantID  = "75093633-1578-497f-890a-d438a74a4127"
	firstName = "given_first_name"
	lastName  = "given_last_name"
	age       = 55

	tenantIsolationConditionWithoutOwnerCheckFmt = "(id IN (SELECT id FROM %s WHERE tenant_id = %s))"
	tenantIsolationConditionWithOwnerCheckFmt    = "(id IN (SELECT id FROM %s WHERE tenant_id = %s AND owner = true))"
	tenantIsolationConditionForBIA               = "(id IN (SELECT id FROM %s WHERE tenant_id = %s AND owner = true) OR owner_id = %s)"
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

var fixApp2 = &App{
	ID:          appID2,
	Name:        appName,
	Description: appDescription,
}

var fixBundle = &Bundle{
	ID:          bundleID,
	Name:        bundleName,
	Description: bundleDescription,
	AppID:       appID,
}

var fixBIA = &BundleInstanceAuth{
	ID:          biaID,
	Name:        biaName,
	Description: biaDescription,
	OwnerID:     tenantID,
	BundleID:    bundleID,
}

var fixWebhook = &Webhook{
	ID:                  whID,
	FormationTemplateID: ftID,
}

// User is a exemplary type to test generic Repositories
type User struct {
	ID        string `db:"id"`
	Tenant    string `db:"tenant_id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Age       int
}

func (a User) GetID() string {
	return a.ID
}

const UserType = resource.Type("UserType")

type UserCollection []User

func (u UserCollection) Len() int {
	return len(u)
}

type App struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

func (a *App) GetID() string {
	return a.ID
}

var appColumns = []string{"id", "name", "description"}

type AppCollection []App

func (a AppCollection) Len() int {
	return len(a)
}

func (a *App) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*App
		TenantID string `db:"tenant_id"`
	}{
		App:      a,
		TenantID: tenant,
	}
}

type Bundle struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	AppID       string `db:"app_id"`
}

func (a *Bundle) GetID() string {
	return a.ID
}

func (a *Bundle) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.Application, a.AppID
}

func (a *Bundle) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Bundle
		TenantID string `db:"tenant_id"`
	}{
		Bundle:   a,
		TenantID: tenant,
	}
}

var bundleColumns = []string{"id", "name", "description", "app_id"}
var webhookColumns = []string{"id", "formation_template_id"}

type Webhook struct {
	ID                  string `db:"id"`
	FormationTemplateID string `db:"formation_template_id"`
}

func (w *Webhook) GetID() string {
	return w.ID
}

func (w *Webhook) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.FormationTemplate, w.FormationTemplateID
}

type BundleInstanceAuth struct {
	ID          string `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	OwnerID     string `db:"owner_id"`
	BundleID    string `db:"bundle_id"`
}

func (a *BundleInstanceAuth) GetID() string {
	return a.ID
}

func (a *BundleInstanceAuth) GetParent(_ resource.Type) (resource.Type, string) {
	return resource.Bundle, a.BundleID
}

func (a *BundleInstanceAuth) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*BundleInstanceAuth
		TenantID string `db:"tenant_id"`
	}{
		BundleInstanceAuth: a,
		TenantID:           tenant,
	}
}

func someError() error {
	return errors.New("some error")
}
