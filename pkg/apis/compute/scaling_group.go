package compute

import "yunion.io/x/onecloud/pkg/apis"

type ScalingGroupCreateInput struct {
	apis.VirtualResourceCreateInput
	apis.EnabledBaseResourceCreateInput

	// description: cloud region id or name
	// required: true
	// example: cr-test-one
	Cloudregion string `json:"cloudregion"`

	//swagger: ignore
	CloudregionId string `json:"cloudregion_id"`

	// description: hypervisor
	// example: kvm
	Hypervisor string `json:"hypervisor"`

	// description: VPC(ID or Name)
	// example: vpc-1234
	Vpc string `json:"vpc"`

	// swagger: ignore
	VpcId string `json:"vpc_id"`

	// description: 多个网络(ID或者Name),
	// example: n-test-one
	Networks []string `json:"networks"`

	// description: 最小实例数
	// example: 0
	MinInstanceNumber int `json:"min_instance_number"`

	// description: 最大实例数
	// example: 10
	MaxInstanceNumber int `json:"max_instance_number"`

	// description: 期望实例数
	// example: 1
	DesireInstanceNumber int `json:"desire_instance_number"`

	// description: 主机模板 id or name
	// example: gt-test-one
	GuestTemplate string `json:"guest_template"`

	// swagger: ignore
	GuestTemplateId string `json:"guest_template_id"`

	// description: 扩容策略
	// enum: balanced
	// required: false
	// example: balanced
	ExpansionPrinciple string `json:"expansion_principle"`

	// description: 缩容策略
	// enum: earliest,latest,config_earliest,config_latest
	// example: latest
	ShrinkPrinciple string `json:"shrink_principle"`

	// description: 检查健康模式
	// enum: normal,loadbalancer
	// example: normal
	HealthCheckMode string `json:"health_check_mode"`

	// description: 健康检查周期
	// example: 300
	HealthCheckCycle int `json:"health_check_cycle"`

	// description: 健康检查缓冲时间
	// example: 180
	HealthCheckGov int `json:"health_check_gov"`

	// description: 负载均衡后端服务器组
	// example: lbg-nihao
	LbBackendGroup string `json:"lb_backend_group"`

	// swagger: ignore
	BackendGroupId string `json:"backend_group_id"`

	// description: 负载均衡后端服务器统一端口
	LoadbalancerBackendPort int `json:"loadbalancer_backend_port"`
}

type ScalingGroupListInput struct {
	apis.VirtualResourceListInput
	RegionalFilterListInput
	LoadbalancerBackendGroupFilterListInput
	VpcFilterListInput
	GroupFilterListInput
	GuestTemplateFilterListInput
	apis.EnabledResourceBaseListInput
	// description: hypervisor
	// example: kvm
	Hypervisor string `json:"hypervisor"`
}

type ScalingGroupDetails struct {
	apis.VirtualResourceDetails
	CloudregionResourceInfo
	LoadbalancerBackendGroupResourceInfo
	VpcResourceInfo
	GroupResourceInfo
	GuestTemplateResourceInfo
	SScalingGroup

	// description: 实例数
	// example: 0
	InstanceNumber int `json:"instance_number"`

	// description: 伸缩策略的数量
	// example: 3
	ScalingPolicyNumber int `json:"scaling_policy_number"`

	// description: 网络ID
	// example: net-12345
	Networks []string `json:"networks"`
}

type ScalingGroupResourceInfo struct {
	// description: 伸缩组名称
	// example: sg-nihao
	ScalingGroup string `json:"scaling_group"`

	// description: 伸缩组ID
	// example: sg-1234
	ScalingGroupId string `json:"scaling_group_id"`
}

type ScalingGroupFilterListInput struct {
	// descirption: 伸缩组 Id or Name
	// example: sg-1234
	ScalingGroup string `json:"scaling_group"`
}

type SGPerformDetachScalingGroupInput struct {
	// description: 伸缩组 Id or Name
	// example: sg-1234
	ScalingGroup string `json:"scaling_group"`

	// description: 是否删除机器
	// example: false
	DeleteServer bool `json:"delete_server"`
}
