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

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/util/reflectutils"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

type SCloudregionResourceBase struct {
	// 归属区域ID
	CloudregionId string `width:"36" charset:"ascii" nullable:"false" list:"user" default:"default" create:"optional" json:"cloudregion_id"`
}

type SCloudregionResourceBaseManager struct{}

func (self *SCloudregionResourceBase) GetRegion() *SCloudregion {
	region, err := CloudregionManager.FetchById(self.CloudregionId)
	if err != nil {
		log.Errorf("failed to find cloudregion %s error: %v", self.CloudregionId, err)
		return nil
	}
	return region.(*SCloudregion)
}

func (self *SCloudregionResourceBase) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) api.CloudregionResourceInfo {
	return api.CloudregionResourceInfo{}
}

func (manager *SCloudregionResourceBaseManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []api.CloudregionResourceInfo {
	rows := make([]api.CloudregionResourceInfo, len(objs))
	regionIds := make([]string, len(objs))
	for i := range objs {
		var base *SCloudregionResourceBase
		reflectutils.FindAnonymouStructPointer(objs[i], &base)
		if base != nil && len(base.CloudregionId) > 0 {
			regionIds[i] = base.CloudregionId
		}
	}
	regions := make(map[string]SCloudregion)
	err := db.FetchStandaloneObjectsByIds(CloudregionManager, regionIds, regions)
	if err != nil {
		log.Errorf("FetchStandaloneObjectsByIds fail %s", err)
		return rows
	}
	for i := range rows {
		if region, ok := regions[regionIds[i]]; ok {
			rows[i] = region.GetRegionInfo()
		} else {
			rows[i] = api.CloudregionResourceInfo{}
		}
	}
	return rows
}

func (manager *SCloudregionResourceBaseManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.RegionalFilterListInput,
) (*sqlchemy.SQuery, error) {
	return managedResourceFilterByRegion(q, query, "", nil)
}

func (manager *SCloudregionResourceBaseManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.RegionalFilterListInput,
) (*sqlchemy.SQuery, error) {
	q, orders, fields := manager.GetOrderBySubQuery(q, userCred, query)
	if len(orders) > 0 {
		q = db.OrderByFields(q, orders, fields)
	}
	return q, nil
}

func (manager *SCloudregionResourceBaseManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	if field == "region" {
		regionQuery := CloudregionManager.Query("name", "id").Distinct().SubQuery()
		q.AppendField(regionQuery.Field("name", field))
		q = q.Join(regionQuery, sqlchemy.Equals(q.Field("cloudregion_id"), regionQuery.Field("id")))
		q.GroupBy(regionQuery.Field("name"))
		return q, nil
	}
	return q, httperrors.ErrNotFound
}

func (manager *SCloudregionResourceBaseManager) GetOrderBySubQuery(
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.RegionalFilterListInput,
) (*sqlchemy.SQuery, []string, []sqlchemy.IQueryField) {
	regionQ := CloudregionManager.Query("id", "name", "city")
	var orders []string
	var fields []sqlchemy.IQueryField
	if db.NeedOrderQuery(manager.GetOrderByFields(query)) {
		subq := regionQ.SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("cloudregion_id"), subq.Field("id")))
		orders = append(orders, query.OrderByRegion, query.OrderByCity)
		fields = append(fields, subq.Field("name"), subq.Field("city"))
	}
	return q, orders, fields
}

func (manager *SCloudregionResourceBaseManager) GetOrderByFields(query api.RegionalFilterListInput) []string {
	return []string{query.OrderByRegion, query.OrderByCity}
}
