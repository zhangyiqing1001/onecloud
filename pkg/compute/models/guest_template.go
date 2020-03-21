// Copyright 2019 Yunion
//
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

package models

import (
	"context"
	"fmt"
	"yunion.io/x/pkg/utils"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/sets"
	"yunion.io/x/sqlchemy"

	"yunion.io/x/onecloud/pkg/apis"
	computeapis "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/cmdline"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/mcclient/auth"
	"yunion.io/x/onecloud/pkg/mcclient/modules"
	"yunion.io/x/onecloud/pkg/util/logclient"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

const (
	IMAGE_TYPE_NORMAL = "normal"
	IMAGE_TYPE_GUEST  = "guest"
)

type SGuestTemplateManager struct {
	db.SSharableVirtualResourceBaseManager
	SCloudregionResourceBaseManager
	SVpcResourceBaseManager
}

type SGuestTemplate struct {
	db.SSharableVirtualResourceBase
	SCloudregionResourceBase
	SVpcResourceBase

	// 虚拟机CPU数量
	VcpuCount int `nullable:"false" default:"1" create:"optional" json:"vcpu_count"`

	// 虚拟机内存大小（MB）
	VmemSize int `nullable:"false" create:"optional" json:"vmem_size"`

	// 虚拟机操作系统类型
	// pattern:Linux|Windows|VMWare
	OsType string `width:"36" charset:"ascii" nullable:"true" create:"optional" json:"os_type"`

	// 镜像类型
	ImageType string `width:"10" charset:"ascii" nullabel:"true" default:"normal" create:"optional" json:"image_type"`

	// 镜像ID
	ImageId string `width:"128" charset:"ascii" create:"optional" json:"image_id"`

	// 虚拟机技术
	Hypervisor string `width:"16" charset:"ascii" default:"kvm" create:"optional" json:"hypervisor"`

	// 其他配置信息
	Content jsonutils.JSONObject `nullable:"false" list:"user" update:"user" create:"optional" json:"content"`
}

var GuestTemplateManager *SGuestTemplateManager

func init() {
	GuestTemplateManager = &SGuestTemplateManager{
		SSharableVirtualResourceBaseManager: db.NewSharableVirtualResourceBaseManager(
			SGuestTemplate{},
			"guesttemplates_tbl",
			"servertemplate",
			"servertemplates",
		),
	}

	GuestTemplateManager.SetVirtualObject(GuestTemplateManager)
}

func (gtm *SGuestTemplateManager) ValidateCreateData(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider,
	query jsonutils.JSONObject,
	input computeapis.GuestTemplateCreateInput,
) (computeapis.GuestTemplateCreateInput, error) {

	var err error
	if input.Content == nil {
		return input, httperrors.NewMissingParameterError("content")
	}

	input.GuestTemplateInput, err = gtm.validateData(ctx, userCred, ownerId, query, input.GuestTemplateInput)
	if err != nil {
		return input, errors.Wrap(err, "gtm.validateData")
	}

	input.SharableVirtualResourceCreateInput, err = gtm.SSharableVirtualResourceBaseManager.ValidateCreateData(ctx, userCred, ownerId, query, input.SharableVirtualResourceCreateInput)
	if err != nil {
		return input, errors.Wrap(err, "SSharableVirtualResourceBaseManager.ValidateCreateData")
	}

	return input, nil
}

func (gt *SGuestTemplate) PostCreate(ctx context.Context, userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data jsonutils.JSONObject) {

	logclient.AddActionLogWithContext(ctx, gt, logclient.ACT_CREATE, nil, userCred, true)
}

func (gt *SGuestTemplate) PostUpdate(ctx context.Context, userCred mcclient.TokenCredential,
	query jsonutils.JSONObject, data jsonutils.JSONObject) {
	logclient.AddActionLogWithContext(ctx, gt, logclient.ACT_UPDATE, nil, userCred, true)
}

func (gtm *SGuestTemplateManager) validateData(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider,
	query jsonutils.JSONObject,
	cinput computeapis.GuestTemplateInput,
) (computeapis.GuestTemplateInput, error) {
	if cinput.Content == nil {
		return cinput, nil
	}
	content := cinput.Content
	// data := cinput.JSON(cinput)
	// not support guest image and guest snapshot for now
	if content.Contains("instance_snapshot_id") {
		return cinput, httperrors.NewInputParameterError(
			"no support for instance snapshot in guest template for now")
	}
	// I don't hope cinput.Content same with data["content"] will change in GuestManager.validateCreateData
	copy := jsonutils.DeepCopy(content).(*jsonutils.JSONDict)
	input, err := GuestManager.validateCreateData(ctx, userCred, ownerId, query, copy)
	if err != nil {
		return cinput, httperrors.NewInputParameterError(err.Error())
	}
	// fill field
	cinput.VmemSize = input.VmemSize
	// data.Add(jsonutils.NewInt(int64(input.VmemSize)), "vmem_size")
	cinput.VcpuCount = input.VcpuCount
	// data.Add(jsonutils.NewInt(int64(input.VcpuCount)), "vcpu_count")
	cinput.OsType = input.OsType
	// data.Add(jsonutils.NewString(input.OsType), "os_type")
	cinput.Hypervisor = input.Hypervisor
	// data.Add(jsonutils.NewString(input.Hypervisor), "hypervisor")
	cinput.InstanceType = input.InstanceType
	cinput.CloudregionId = input.PreferRegion

	// fill vpc
	if len(input.Networks) != 0 && len(input.Networks[0].Network) != 0 {
		model, err := WireManager.FetchByIdOrName(userCred, input.Networks[0].Wire)
		if err != nil {
			return cinput, errors.Wrap(err, "WireManger.FetchByIdOrName")
		}
		cinput.VpcId = model.(*SWire).VpcId
	}

	if len(input.GuestImageID) > 0 {
		cinput.ImageType = IMAGE_TYPE_GUEST
		cinput.ImageId = input.GuestImageID
		// data.Add(jsonutils.NewString(IMAGE_TYPE_GUEST), "image_type")
		// data.Add(jsonutils.NewString(input.GuestImageID), "image_id")
	} else {
		cinput.ImageType = input.GuestImageID
		cinput.ImageId = input.Disks[0].ImageId // if input.Didks is empty???
		// data.Add(jsonutils.NewString(input.Disks[0].ImageId), "image_id")
		// data.Add(jsonutils.NewString(IMAGE_TYPE_NORMAL), "image_type")
	}

	// hide some properties
	contentDict := content.(*jsonutils.JSONDict)
	contentDict.Remove("name")
	contentDict.Remove("generate_name")
	// "__count__" was converted to "count" by apigateway
	contentDict.Remove("count")
	contentDict.Remove("project_id")

	cinput.Content = contentDict
	// data.Add(contentDict, "content")
	return cinput, nil
}

func (gt *SGuestTemplate) ValidateUpdateData(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	input computeapis.GuestTemplateUpdateInput,
) (computeapis.GuestTemplateUpdateInput, error) {
	var err error
	input.GuestTemplateInput, err = GuestTemplateManager.validateData(ctx, userCred, gt.GetOwnerId(), query, input.GuestTemplateInput)
	if err != nil {
		return input, errors.Wrap(err, "GuestTemplateManager.validateData")
	}
	input.SharableVirtualResourceBaseUpdateInput, err = gt.SSharableVirtualResourceBase.ValidateUpdateData(ctx, userCred, query, input.SharableVirtualResourceBaseUpdateInput)
	if err != nil {
		return input, errors.Wrap(err, "SSharableVirtualResourceBase.ValidateUpdateData")
	}
	return input, nil
}

func (manager *SGuestTemplateManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []computeapis.GuestTemplateDetails {
	rows := make([]computeapis.GuestTemplateDetails, len(objs))

	virtRows := manager.SSharableVirtualResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	crRows := manager.SCloudregionResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	vpcRows := manager.SVpcResourceBaseManager.FetchCustomizeColumns(ctx, userCred, query, objs, fields, isList)
	for i := range rows {
		rows[i] = computeapis.GuestTemplateDetails{
			SharableVirtualResourceDetails: virtRows[i],
			CloudregionResourceInfo:        crRows[i],
			VpcResourceInfo:                vpcRows[i],
		}
		rows[i], _ = objs[i].(*SGuestTemplate).getMoreDetails(ctx, userCred, rows[i])
	}

	return rows
}

func (gt *SGuestTemplate) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) (computeapis.GuestTemplateDetails, error) {
	return computeapis.GuestTemplateDetails{}, nil
}

func (gt *SGuestTemplate) getMoreDetails(ctx context.Context, userCred mcclient.TokenCredential,
	out computeapis.GuestTemplateDetails) (computeapis.GuestTemplateDetails, error) {

	input, err := cmdline.FetchServerCreateInputByJSON(gt.Content)
	if err != nil {
		return out, err
	}
	configInfo := computeapis.GuestTemplateConfigInfo{}
	if len(input.PreferRegion) != 0 {
		region := CloudregionManager.FetchRegionById(input.PreferRegion)
		if region != nil {
			input.PreferRegion = region.GetName()
		}
		configInfo.Region = input.PreferRegion

	}
	if len(input.PreferZone) != 0 {
		zone := ZoneManager.FetchZoneById(input.PreferZone)
		if zone != nil {
			input.PreferZone = zone.GetName()
		}
		configInfo.Zone = input.PreferZone
	}
	configInfo.Hypervisor = gt.Hypervisor
	configInfo.OsType = gt.OsType

	// sku deal
	if len(input.InstanceType) > 0 {
		skuOutput := computeapis.GuestTemplateSku{}
		provider := GetDriver(gt.Hypervisor).GetProvider()
		sku, err := ServerSkuManager.FetchSkuByNameAndProvider(input.InstanceType, provider, true)
		if err != nil {
			skuOutput.Name = input.InstanceType
			skuOutput.MemorySizeMb = gt.VmemSize
			skuOutput.CpuCoreCount = gt.VcpuCount
		} else {
			skuOutput.Name = sku.Name
			skuOutput.MemorySizeMb = sku.MemorySizeMB
			skuOutput.CpuCoreCount = sku.CpuCoreCount
			skuOutput.InstanceTypeCategory = sku.InstanceTypeCategory
			skuOutput.InstanceTypeFamily = sku.InstanceTypeFamily
		}
		configInfo.Sku = skuOutput
	}

	// disk deal
	disks := make([]computeapis.GuestTemplateDisk, len(input.Disks))
	for i := range input.Disks {
		disks[i] = computeapis.GuestTemplateDisk{
			Backend:  input.Disks[i].Backend,
			DiskType: input.Disks[i].DiskType,
			Index:    input.Disks[i].Index,
			SizeMb:   input.Disks[i].SizeMb,
		}
	}
	configInfo.Disks = disks

	// keypair
	if len(input.KeypairId) > 0 {
		model, err := KeypairManager.FetchById(input.KeypairId)
		if err == nil {
			keypair := model.(*SKeypair)
			configInfo.Keypair = keypair.GetName()
		}
	}

	// network
	if len(input.Networks) > 0 {
		networkList := make([]computeapis.GuestTemplateNetwork, 0, len(input.Networks))
		networkIdList := make([]string, len(input.Networks))
		for i := range input.Networks {
			networkIdList[i] = input.Networks[i].Network
		}
		networkSet := sets.NewString(networkIdList...)

		wireQuery := WireManager.Query("id", "name", "vpc_id").SubQuery()
		vpcQuery := VpcManager.Query("id", "name").SubQuery()
		q := NetworkManager.Query("id", "name", "wire_id", "guest_ip_start", "guest_ip_end", "vlan_id")
		if len(networkIdList) == 1 {
			q = q.Equals("id", networkIdList[0])
		} else {
			q = q.In("id", networkIdList)
		}
		q = q.LeftJoin(wireQuery, sqlchemy.Equals(q.Field("wire_id"), wireQuery.Field("id")))
		q = q.LeftJoin(vpcQuery, sqlchemy.Equals(wireQuery.Field("vpc_id"), vpcQuery.Field("id")))
		q = q.AppendField(vpcQuery.Field("id", "vpc_id"), vpcQuery.Field("name", "vpc_name"))
		q.All(&networkList)

		for _, p := range networkList {
			if networkSet.Has(p.ID) {
				networkSet.Delete(p.ID)
			}
		}

		// some specified network
		for _, id := range networkSet.UnsortedList() {
			networkList = append(networkList, computeapis.GuestTemplateNetwork{ID: id})
		}

		configInfo.Nets = networkList
	}

	// secgroup
	if len(input.SecgroupId) > 0 {
		secgroup := SecurityGroupManager.FetchSecgroupById(input.SecgroupId)
		if secgroup != nil {
			input.SecgroupId = secgroup.GetName()
		}
		configInfo.Secgroup = input.SecgroupId
	}

	// isolatedDevices
	if input.IsolatedDevices != nil && len(input.IsolatedDevices) != 0 {
		configInfo.IsolatedDeviceConfig = make([]computeapis.IsolatedDeviceConfig, len(input.IsolatedDevices))
		for i := range configInfo.IsolatedDeviceConfig {
			configInfo.IsolatedDeviceConfig[i] = *input.IsolatedDevices[i]
		}
	}

	// fill image info
	switch gt.ImageType {
	case IMAGE_TYPE_NORMAL:
		image, err := CachedimageManager.getImageInfo(ctx, userCred, gt.ImageId, false)
		if err == nil {
			configInfo.Image = image.Name
		} else {
			configInfo.Image = gt.ImageId
		}
	case IMAGE_TYPE_GUEST:
		s := auth.GetSession(ctx, userCred, options.Options.Region, "")
		ret, err := modules.GuestImages.Get(s, gt.ImageId, jsonutils.JSONNull)
		if err != nil || !ret.Contains("id") {
			configInfo.Image = gt.ImageId
		} else {
			name, _ := ret.GetString("id")
			configInfo.Image = name
		}
	default:
		// no arrivals
	}

	// reset_password
	if input.ResetPassword == nil {
		configInfo.ResetPassword = false
	} else {
		configInfo.ResetPassword = *input.ResetPassword
	}

	out.ConfigInfo = configInfo
	return out, nil
}

func (gt *SGuestTemplate) PerformPublic(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	data apis.PerformPublicInput,
) (jsonutils.JSONObject, error) {

	// image, network, secgroup, instancegroup
	input, err := cmdline.FetchServerCreateInputByJSON(gt.Content)
	if err != nil {
		return nil, errors.Wrap(err, "fail to convert content of guest template to ServerCreateInput")
	}

	// check for below private resource in the guest template
	privateResource := map[string]int{
		"keypair":           len(input.KeypairId),
		"instance group":    len(input.InstanceGroupIds),
		"instance snapshot": len(input.InstanceSnapshotId),
	}
	for k, v := range privateResource {
		if v > 0 {
			return nil, gt.genForbiddenError(k, "", "")
		}
	}

	targetScopeStr := data.Scope
	targetScope := rbacutils.String2ScopeDefault(targetScopeStr, rbacutils.ScopeSystem)

	// check if secgroup is public
	if len(input.SecgroupId) > 0 {
		model, err := SecurityGroupManager.FetchByIdOrName(userCred, input.SecgroupId)
		if err != nil {
			return nil, httperrors.NewResourceNotFoundError("there is no such secgroup %s descripted by guest template",
				input.SecgroupId)
		}
		secgroup := model.(*SSecurityGroup)
		sgScope := rbacutils.String2Scope(secgroup.PublicScope)
		if !secgroup.IsPublic || !sgScope.HigherEqual(targetScope) {
			return nil, gt.genForbiddenError("security group", input.SecgroupId, string(targetScope))
		}
	}

	// check if networks is public
	if len(input.Networks) > 0 {
		for i := range input.Networks {
			str := input.Networks[i].Network
			model, err := NetworkManager.FetchByIdOrName(userCred, str)
			if err != nil {
				return nil, httperrors.NewResourceNotFoundError(
					"there is no such secgroup %s descripted by guest template")
			}
			network := model.(*SNetwork)
			netScope := rbacutils.String2Scope(network.PublicScope)
			if !network.IsPublic || !netScope.HigherEqual(targetScope) {
				return nil, gt.genForbiddenError("network", str, string(targetScope))
			}
		}
	}

	// check if image is public
	var (
		isPublic    bool
		publicScope string
	)
	switch gt.ImageType {
	case IMAGE_TYPE_NORMAL:
		image, err := CachedimageManager.GetImageById(ctx, userCred, gt.ImageId, false)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to fetch image %s descripted by guest template", gt.ImageId)
		}
		isPublic, publicScope = image.IsPublic, image.PublicScope
	case IMAGE_TYPE_GUEST:
		s := auth.GetSession(ctx, userCred, options.Options.Region, "")
		ret, err := modules.GuestImages.Get(s, gt.ImageId, jsonutils.JSONNull)
		if err != nil {
			return nil, errors.Wrapf(err, "fail to fetch guest image %s descripted by guest template", gt.ImageId)
		}
		isPublic = jsonutils.QueryBoolean(ret, "is_public", false)
		publicScope, _ = ret.GetString("public_scope")
	default:
		//no arrivals
	}
	igScope := rbacutils.String2Scope(publicScope)
	if !isPublic || !igScope.HigherEqual(targetScope) {
		return nil, gt.genForbiddenError("image", "", string(targetScope))
	}

	return gt.SSharableVirtualResourceBase.PerformPublic(ctx, userCred, query, data)
}

func (gt *SGuestTemplate) genForbiddenError(resourceName, resourceStr, scope string) error {
	var msg string
	if len(resourceStr) == 0 {
		msg = fmt.Sprintf("the %s in guest template is not a public resource", resourceName)
	} else {
		msg = fmt.Sprintf("the %s '%s' in guest template is not a public resource", resourceName, resourceStr)
	}
	if len(scope) > 0 {
		msg += fmt.Sprintf(" in %s scope", scope)
	}
	return httperrors.NewForbiddenError(msg)
}

func (gt *SGuestTemplate) ValidateDeleteCondition(ctx context.Context) error {
	q := ServiceCatalogManager.Query("name").Equals("guest_template_id", gt.Id)
	names := make([]struct{ Name string }, 0, 1)
	err := q.All(&names)
	if err != nil {
		return errors.Wrap(err, "SQuery.All")
	}
	if len(names) > 0 {
		return httperrors.NewForbiddenError("guest template %s used by service catalog %s", gt.Id, names[0].Name)
	}
	return nil
}

// 主机模板列表
func (manager *SGuestTemplateManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	input computeapis.GuestTemplateListInput,
) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SSharableVirtualResourceBaseManager.ListItemFilter(ctx, q, userCred, input.SharableVirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SSharableVirtualResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SCloudregionResourceBaseManager.ListItemFilter(ctx, q, userCred, input.RegionalFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SCloudregionResourceBaseManager.ListItemFilter")
	}
	q, err = manager.SVpcResourceBaseManager.ListItemFilter(ctx, q, userCred, input.VpcFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVpcResourceBaseManager.ListItemFilter")
	}
	return q, nil
}

func (manager *SGuestTemplateManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	input computeapis.GuestTemplateListInput,
) (*sqlchemy.SQuery, error) {
	var err error

	q, err = manager.SSharableVirtualResourceBaseManager.OrderByExtraFields(ctx, q, userCred, input.SharableVirtualResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SSharableVirtualResourceBaseManager.OrderByExtraFields")
	}
	q, err = manager.SCloudregionResourceBaseManager.OrderByExtraFields(ctx, q, userCred, input.RegionalFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SCloudregionResourceBaseManager.OrderByExtraFields")
	}
	q, err = manager.SVpcResourceBaseManager.OrderByExtraFields(ctx, q, userCred, input.VpcFilterListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SVpcResourceBaseManager.OrderByExtraFields")
	}
	return q, nil
}

func (manager *SGuestTemplateManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	var err error

	q, err = manager.SSharableVirtualResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	q, err = manager.SCloudregionResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}
	q, err = manager.SVpcResourceBaseManager.QueryDistinctExtraField(q, field)
	if err == nil {
		return q, nil
	}

	return q, httperrors.ErrNotFound
}

type SGuestTemplateValidate struct {
	Hypervisor    string
	CloudregionId string
	VpcId         string
	NetworkIds    []string
}

func (gt *SGuestTemplate) Validate(ctx context.Context, userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider, stv SGuestTemplateValidate) (bool, string) {
	if stv.Hypervisor != "" && gt.Hypervisor != stv.Hypervisor {
		return false, fmt.Sprintf("GuestTemplate has mismatched hypervisor, need %s but %s", stv.Hypervisor, gt.Hypervisor)
	}
	if stv.CloudregionId != "" && gt.CloudregionId != stv.CloudregionId {
		return false, fmt.Sprintf("GuestTemplate has mismatched cloudregion, need %s but %s", stv.CloudregionId, gt.CloudregionId)
	}
	if stv.VpcId != "" && gt.VpcId != "" && stv.VpcId != gt.VpcId {
		return false, fmt.Sprintf("GuestTemplate has mismatched vpc, need %s bu %s", stv.VpcId, gt.VpcId)
	}

	// check networks
	input, err := GuestManager.validateCreateData(ctx, userCred, ownerId, jsonutils.NewDict(), gt.Content.(*jsonutils.JSONDict))
	if err != nil {
		return false, err.Error()
	}
	for i := range input.Networks {
		if !utils.IsInStringArray(input.Networks[i].Network, stv.NetworkIds) {
			return false, fmt.Sprintf("GuestTemplate' network '%s' not in networks '%s'", input.Networks[i].Network,
				stv.NetworkIds)
		}
	}
	return true, ""
}
