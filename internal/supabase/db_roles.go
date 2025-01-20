/*
Copyright 2025 Peter Kurfer.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
