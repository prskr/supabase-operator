package supabase

import "strings"

type DBRole string

func (r DBRole) String() string {
	return string(r)
}

func (r DBRole) K8sString() string {
	return strings.ReplaceAll(r.String(), "_", "-")
}

func (r DBRole) Bytes() []byte {
	s := string(r)
	return []byte(s)
}

const (
	DBRoleAuthenticator  DBRole = "authenticator"
	DBRoleAuthAdmin      DBRole = "supabase_auth_admin"
	DBRoleFunctionsAdmin DBRole = "supabase_functions_admin"
	DBRoleStorageAdmin   DBRole = "supabase_storage_admin"
	DBRoleSupabaseAdmin  DBRole = "supabase_admin"
)
