// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by model-api-gen. DO NOT EDIT.

package identity

import (
	time "time"

	jsonutils "yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/apis"
)

// SAssignment is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SAssignment.
type SAssignment struct {
	apis.SResourceBase
	Type      string `json:"type"`
	ActorId   string `json:"actor_id"`
	TargetId  string `json:"target_id"`
	RoleId    string `json:"role_id"`
	Inherited *bool  `json:"inherited,omitempty"`
}

// SConfigOption is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SConfigOption.
type SConfigOption struct {
	apis.SResourceBase
	ResType string               `json:"res_type"`
	ResId   string               `json:"res_id"`
	Group   string               `json:"group"`
	Option  string               `json:"option"`
	Value   jsonutils.JSONObject `json:"value"`
}

// SCredential is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SCredential.
type SCredential struct {
	apis.SStandaloneResourceBase
	UserId        string              `json:"user_id"`
	ProjectId     string              `json:"project_id"`
	Type          string              `json:"type"`
	KeyHash       string              `json:"key_hash"`
	Extra         *jsonutils.JSONDict `json:"extra"`
	EncryptedBlob string              `json:"encrypted_blob"`
	Enabled       *bool               `json:"enabled,omitempty"`
}

// SDomain is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SDomain.
type SDomain struct {
	apis.SStandaloneResourceBase
	Extra    *jsonutils.JSONDict `json:"extra"`
	Enabled  *bool               `json:"enabled,omitempty"`
	IsDomain *bool               `json:"is_domain,omitempty"`
	DomainId string              `json:"domain_id"`
	ParentId string              `json:"parent_id"`
}

// SEnabledIdentityBaseResource is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SEnabledIdentityBaseResource.
type SEnabledIdentityBaseResource struct {
	SIdentityBaseResource
	Enabled *bool `json:"enabled,omitempty"`
}

// SEndpoint is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SEndpoint.
type SEndpoint struct {
	apis.SStandaloneResourceBase
	LegacyEndpointId string              `json:"legacy_endpoint_id"`
	Interface        string              `json:"interface"`
	ServiceId        string              `json:"service_id"`
	Url              string              `json:"url"`
	Extra            *jsonutils.JSONDict `json:"extra"`
	Enabled          *bool               `json:"enabled,omitempty"`
	RegionId         string              `json:"region_id"`
}

// SFederatedUser is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SFederatedUser.
type SFederatedUser struct {
	apis.SResourceBase
	Id          int    `json:"id"`
	UserId      string `json:"user_id"`
	IdpId       string `json:"idp_id"`
	ProtocolId  string `json:"protocol_id"`
	UniqueId    string `json:"unique_id"`
	DisplayName string `json:"display_name"`
}

// SIdentityBaseResource is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SIdentityBaseResource.
type SIdentityBaseResource struct {
	apis.SStandaloneResourceBase
	apis.SDomainizedResourceBase
	Extra *jsonutils.JSONDict `json:"extra"`
}

// SIdentityProvider is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SIdentityProvider.
type SIdentityProvider struct {
	apis.SEnabledStatusStandaloneResourceBase
	Driver              string    `json:"driver"`
	Template            string    `json:"template"`
	TargetDomainId      string    `json:"target_domain_id"`
	AutoCreateProject   *bool     `json:"auto_create_project,omitempty"`
	ErrorCount          int       `json:"error_count"`
	SyncStatus          string    `json:"sync_status"`
	LastSync            time.Time `json:"last_sync"`
	LastSyncEndAt       time.Time `json:"last_sync_end_at"`
	SyncIntervalSeconds int       `json:"sync_interval_seconds"`
}

// SIdmapping is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SIdmapping.
type SIdmapping struct {
	apis.SResourceBase
	PublicId    string `json:"public_id"`
	IdpId       string `json:"idp_id"`
	IdpEntityId string `json:"idp_entity_id"`
	EntityType  string `json:"entity_type"`
}

// SLocalUser is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SLocalUser.
type SLocalUser struct {
	apis.SResourceBase
	Id              int       `json:"id"`
	UserId          string    `json:"user_id"`
	DomainId        string    `json:"domain_id"`
	Name            string    `json:"name"`
	FailedAuthCount int       `json:"failed_auth_count"`
	FailedAuthAt    time.Time `json:"failed_auth_at"`
}

// SPassword is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SPassword.
type SPassword struct {
	apis.SResourceBase
	Id           int       `json:"id"`
	LocalUserId  int       `json:"local_user_id"`
	Password     string    `json:"password"`
	ExpiresAt    time.Time `json:"expires_at"`
	SelfService  bool      `json:"self_service"`
	PasswordHash string    `json:"password_hash"`
	CreatedAtInt int64     `json:"created_at_int"`
	ExpiresAtInt int64     `json:"expires_at_int"`
}

// SPolicy is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SPolicy.
type SPolicy struct {
	SEnabledIdentityBaseResource
	apis.SSharableBaseResource
	Type string               `json:"type"`
	Blob jsonutils.JSONObject `json:"blob"`
}

// SRegion is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SRegion.
type SRegion struct {
	apis.SStandaloneResourceBase
	ParentRegionId string              `json:"parent_region_id"`
	Extra          *jsonutils.JSONDict `json:"extra"`
}

// SRole is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SRole.
type SRole struct {
	SIdentityBaseResource
	apis.SSharableBaseResource
}

// SService is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SService.
type SService struct {
	apis.SStandaloneResourceBase
	Type    string              `json:"type"`
	Enabled *bool               `json:"enabled,omitempty"`
	Extra   *jsonutils.JSONDict `json:"extra"`
}

// SUsergroupMembership is an autogenerated struct via yunion.io/x/onecloud/pkg/keystone/models.SUsergroupMembership.
type SUsergroupMembership struct {
	apis.SResourceBase
	UserId  string `json:"user_id"`
	GroupId string `json:"group_id"`
}
