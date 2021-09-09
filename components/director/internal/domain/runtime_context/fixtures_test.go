package runtimectx_test

import (
	"fmt"
	"regexp"
)

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
