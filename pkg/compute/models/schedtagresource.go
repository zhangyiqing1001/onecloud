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
	"database/sql"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/reflectutils"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

type SSchedtagResourceBase struct {
	// 归属调度标签ID
	SchedtagId string `width:"36" charset:"ascii" nullable:"false" list:"user" create:"required" update:"user" json:"schedtag_id"`
}

type SSchedtagResourceBaseManager struct{}

func (self *SSchedtagResourceBase) GetSchedtag() *SSchedtag {
	obj, err := SchedtagManager.FetchById(self.SchedtagId)
	if err != nil {
		log.Errorf("fail to fetch sched tag by id %s", err)
		return nil
	}
	return obj.(*SSchedtag)
}

func (self *SSchedtagResourceBase) GetExtraDetails(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	isList bool,
) api.SchedtagResourceInfo {
	return api.SchedtagResourceInfo{}
}

func (manager *SSchedtagResourceBaseManager) FetchCustomizeColumns(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	query jsonutils.JSONObject,
	objs []interface{},
	fields stringutils2.SSortedStrings,
	isList bool,
) []api.SchedtagResourceInfo {
	rows := make([]api.SchedtagResourceInfo, len(objs))
	schedTagIds := make([]string, len(objs))
	for i := range objs {
		var base *SSchedtagResourceBase
		err := reflectutils.FindAnonymouStructPointer(objs[i], &base)
		if err != nil {
			log.Errorf("Cannot find SSchedtagResourceBase in object %s", objs[i])
			continue
		}
		schedTagIds[i] = base.SchedtagId
	}
	tags := make(map[string]SSchedtag)
	err := db.FetchStandaloneObjectsByIds(SchedtagManager, schedTagIds, tags)
	if err != nil {
		log.Errorf("FetchStandaloneObjectsByIds fail %s", err)
		return rows
	}
	for i := range rows {
		rows[i] = api.SchedtagResourceInfo{}
		if tag, ok := tags[schedTagIds[i]]; ok {
			rows[i].Schedtag = tag.Name
			rows[i].ResourceType = tag.ResourceType
		}
	}
	return rows
}

func (manager *SSchedtagResourceBaseManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.SchedtagFilterListInput,
) (*sqlchemy.SQuery, error) {
	if len(query.Schedtag) > 0 {
		tagObj, err := SchedtagManager.FetchByIdOrName(userCred, query.Schedtag)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, httperrors.NewResourceNotFoundError2(SchedtagManager.Keyword(), query.Schedtag)
			} else {
				return nil, errors.Wrap(err, "SchedtagManager.FetchByIdOrName")
			}
		}
		q = q.Equals("schedtag_id", tagObj.GetId())
	}
	return q, nil
}

func (manager *SSchedtagResourceBaseManager) OrderByExtraFields(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.SchedtagFilterListInput,
) (*sqlchemy.SQuery, error) {
	q, orders, fields := manager.GetOrderBySubQuery(q, userCred, query)
	if len(orders) > 0 {
		q = db.OrderByFields(q, orders, fields)
	}
	return q, nil
}

func (manager *SSchedtagResourceBaseManager) QueryDistinctExtraField(q *sqlchemy.SQuery, field string) (*sqlchemy.SQuery, error) {
	if field == "schedtag" {
		tagQuery := SchedtagManager.Query("name", "id").Distinct().SubQuery()
		q.AppendField(tagQuery.Field("name", field))
		q = q.Join(tagQuery, sqlchemy.Equals(q.Field("schedtag_id"), tagQuery.Field("id")))
		q.GroupBy(tagQuery.Field("name"))
		return q, nil
	}
	return q, httperrors.ErrNotFound
}

func (manager *SSchedtagResourceBaseManager) GetOrderBySubQuery(
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.SchedtagFilterListInput,
) (*sqlchemy.SQuery, []string, []sqlchemy.IQueryField) {
	tagQ := SchedtagManager.Query("id", "name", "resource_type")
	var orders []string
	var fields []sqlchemy.IQueryField
	if db.NeedOrderQuery(manager.GetOrderByFields(query)) {
		subq := tagQ.SubQuery()
		q = q.LeftJoin(subq, sqlchemy.Equals(q.Field("schedtag_id"), subq.Field("id")))
		orders = append(orders, query.OrderBySchedtag, query.OrderByResourceType)
		fields = append(fields, subq.Field("name"), subq.Field("resource_type"))
	}
	return q, orders, fields
}

func (manager *SSchedtagResourceBaseManager) GetOrderByFields(query api.SchedtagFilterListInput) []string {
	return []string{query.OrderBySchedtag, query.OrderByResourceType}
}
