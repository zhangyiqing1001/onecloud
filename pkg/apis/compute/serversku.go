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

package compute

import "yunion.io/x/onecloud/pkg/apis"

type ServerSkuCreateInput struct {
	apis.StatusStandaloneResourceCreateInput

	// 区域名称或Id,建议使用Id
	// default: default
	Cloudregion string `json:"cloudregion"`
	// swagger:ignore
	CloudregionId string

	// 可用区名称或Id, 建议使用Id
	// required: false
	Zone string `json:"zone"`
	// swagger:ignore
	ZoneId string

	// 是否启用
	// default: true
	Enabled *bool `json:"enabled"`

	// swagger:ignore
	InstanceTypeFamily string

	// 套餐类型
	//
	//
	//
	// | instance_type_category	|	说明	|
	// |	-----				|	---		|
	// |general_purpose			|通用型		|
	// |burstable				|突发性能型		|
	// |compute_optimized		|计算优化型		|
	// |memory_optimized		|内存优化型		|
	// |storage_optimized		|存储IO优化型		|
	// |hardware_accelerated	|硬件加速型		|
	// |high_storage			|高存储型		|
	// |high_memory				|高内存型		|
	// default: general_purpose
	InstanceTypeCategory string `json:"instance_type_category"`

	// swagger:ignore
	LocalCategory string

	// 预付费状态
	// default: available
	PrepaidStatus string `json:"prepaid_status"`
	// 后付费状态
	// default: available
	PostpaidStatus string `json:"postpaid_status"`

	// Cpu核数
	// minimum: 1
	// maximum: 256
	// required: true
	CpuCoreCount int64 `json:"cpu_core_count"`

	// 内存大小
	// minimum: 512
	// maximum: 524288
	// required: true
	MemorySizeMB int64 `json:"memory_size_mb"`

	// swagger:ignore
	OsName string

	// swagger:ignore
	SysDiskResizable *bool

	// swagger:ignore
	SysDiskType string

	// swagger:ignore
	SysDiskMinSizeGB int

	// swagger:ignore
	SysDiskMaxSizeGB int

	// swagger:ignore
	AttachedDiskType string

	// swagger:ignore
	AttachedDiskSizeGB int

	// swagger:ignore
	AttachedDiskCount int

	// swagger:ignore
	DataDiskTypes string

	// swagger:ignore
	DataDiskMaxCount int

	// swagger:ignore
	NicType string

	// swagger:ignore
	NicMaxCount int

	// swagger:ignore
	GpuAttachable *bool

	// swagger:ignore
	GpuSpec string

	// swagger:ignore
	GpuCount int

	// swagger:ignore
	GpuMaxCount int

	// swagger:ignore
	Provider string
}

type ServerSkuDetails struct {
	apis.StatusStandaloneResourceDetails

	ZoneResourceInfo

	SServerSku

	// 绑定云主机数量
	TotalGuestCount int `json:"total_guest_count"`
}
