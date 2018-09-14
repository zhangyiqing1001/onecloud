package aliyun

import (
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
)

type ImageStatusType string

const (
	ImageStatusCreating     ImageStatusType = "Creating"
	ImageStatusAvailable    ImageStatusType = "Available"
	ImageStatusUnAvailable  ImageStatusType = "UnAvailable"
	ImageStatusCreateFailed ImageStatusType = "CreateFailed"
)

type ImageOwnerType string

const (
	ImageOwnerSystem      ImageOwnerType = "system"
	ImageOwnerSelf        ImageOwnerType = "self"
	ImageOwnerOthers      ImageOwnerType = "others"
	ImageOwnerMarketplace ImageOwnerType = "marketplace"
)

type ImageUsageType string

const (
	ImageUsageInstance ImageUsageType = "instance"
	ImageUsageNone     ImageUsageType = "none"
)

type SImage struct {
	storageCache *SStoragecache

	Architecture         string
	CreationTime         time.Time
	Description          string
	ImageId              string
	ImageName            string
	OSName               string
	OSType               string
	ImageOwnerAlias      ImageOwnerType
	IsSupportCloudinit   bool
	IsSupportIoOptimized bool
	Platform             string
	Size                 int
	Status               ImageStatusType
	Usage                string
}

func (self *SImage) GetMetadata() *jsonutils.JSONDict {
	data := jsonutils.NewDict()
	if len(self.Architecture) > 0 {
		data.Add(jsonutils.NewString(self.Architecture), "os_arch")
	}
	if len(self.OSType) > 0 {
		data.Add(jsonutils.NewString(self.OSType), "os_name")
	}
	if len(self.Platform) > 0 {
		data.Add(jsonutils.NewString(self.Platform), "os_distribution")
	}
	if len(self.OSName) > 0 {
		data.Add(jsonutils.NewString(self.OSName), "os_version")
	}
	return data
}

func (self *SImage) GetId() string {
	return self.ImageId
}

func (self *SImage) GetName() string {
	return self.ImageName
}

func (self *SImage) IsEmulated() bool {
	return false
}

func (self *SImage) Delete() error {
	return self.storageCache.region.DeleteImage(self.ImageId)
}

func (self *SImage) GetGlobalId() string {
	return fmt.Sprintf("%s-%s")
}

func (self *SImage) GetIStoragecache() cloudprovider.ICloudStoragecache {
	return self.storageCache
}

func (self *SImage) GetStatus() string {
	switch self.Status {
	case ImageStatusCreating:
		return models.IMAGE_STATUS_QUEUED
	case ImageStatusAvailable:
		return models.IMAGE_STATUS_ACTIVE
	case ImageStatusUnAvailable:
		return models.IMAGE_STATUS_DELETED
	case ImageStatusCreateFailed:
		return models.IMAGE_STATUS_KILLED
	default:
		return models.IMAGE_STATUS_KILLED
	}
}

func (self *SImage) Refresh() error {
	new, err := self.storageCache.region.GetImage(self.ImageId)
	if err != nil {
		return err
	}
	return jsonutils.Update(self, new)
}

type ImageExportTask struct {
	ImageId  string
	RegionId string
	// RequestId string
	TaskId string
}

func (self *SRegion) ExportImage(imageId string, bucket *oss.Bucket) (*ImageExportTask, error) {
	params := make(map[string]string)
	params["RegionId"] = self.RegionId
	params["ImageId"] = imageId
	params["OssBucket"] = bucket.BucketName
	params["OssPrefix"] = fmt.Sprintf("%sexport", strings.Replace(imageId, "-", "", -1))

	if body, err := self.ecsRequest("ExportImage", params); err != nil {
		return nil, err
	} else {
		result := ImageExportTask{}
		if err := body.Unmarshal(&result); err != nil {
			log.Errorf("unmarshal result error %s", err)
			return nil, err
		}
		return &result, nil
	}
}

// {"ImageId":"m-j6c1qlpa7oebbg1n2k60","RegionId":"cn-hongkong","RequestId":"F8B2F6A1-F6AA-4C92-A54C-C4A309CF811F","TaskId":"t-j6c1qlpa7oebbg1rcl9t"}

type ImageImportTask struct {
	ImageId  string
	RegionId string
	// RequestId string
	TaskId string
}


func (self *SRegion) ImportImage(name string, osArch string, osType string, osDist string, bucket string, key string) (*ImageImportTask, error) {
	params := make(map[string]string)
	params["RegionId"] = self.RegionId
	params["ImageName"] = name
	params["Platform"] =  osDist // "Others Linux"
	params["OSType"] = osType // "linux"
	params["Architecture"] = osArch // "x86_64"
	params["DiskDeviceMapping.1.OSSBucket"] = bucket
	params["DiskDeviceMapping.1.OSSObject"] = key

	log.Debugf("Upload image with params %#v", params)

	body, err := self.ecsRequest("ImportImage", params)
	if err != nil {
		log.Errorf("ImportImage fail %s", err)
		return nil, err
	}

	log.Infof("%s", body)
	result := ImageImportTask{}
	err = body.Unmarshal(&result)
	if err != nil {
		log.Errorf("unmarshal result error %s", err)
		return nil, err
	}

	return &result, nil
}

func (self *SRegion) GetImage(imageId string) (*SImage, error) {
	images, _, err := self.GetImages("", ImageOwnerSelf, []string{imageId}, "", 0, 1)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("image %s not found", imageId)
	}
	return &images[0], nil
}

func (self *SRegion) GetImageByName(name string) (*SImage, error) {
	images, _, err := self.GetImages("", ImageOwnerSelf, nil, name, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, cloudprovider.ErrNotFound
	}
	return &images[0], nil
}

func (self *SRegion) GetImageStatus(imageId string) (ImageStatusType, error) {
	image, err := self.GetImage(imageId)
	if err != nil {
		return "", err
	}
	return image.Status, nil
}

func (self *SRegion) GetImages(status ImageStatusType, owner ImageOwnerType, imageId []string, name string, offset int, limit int) ([]SImage, int, error) {
	if limit > 50 || limit <= 0 {
		limit = 50
	}
	params := make(map[string]string)
	params["RegionId"] = self.RegionId
	params["PageSize"] = fmt.Sprintf("%d", limit)
	params["PageNumber"] = fmt.Sprintf("%d", (offset/limit)+1)

	if len(status) > 0 {
		params["Status"] = string(status)
	} else {
		params["Status"] = "Creating,Available,UnAvailable,CreateFailed"
	}
	if imageId != nil && len(imageId) > 0 {
		params["ImageId"] = strings.Join(imageId, ",")
	}
	if len(owner) > 0 {
		params["ImageOwnerAlias"] = string(owner)
	}

	if len(name) > 0 {
		params["ImageName"] = name
	}

	// log.Debugf("%s", params)

	body, err := self.ecsRequest("DescribeImages", params)
	if err != nil {
		log.Errorf("DescribeImages fail %s", err)
		return nil, 0, err
	}

	images := make([]SImage, 0)
	err = body.Unmarshal(&images, "Images", "Image")
	if err != nil {
		log.Errorf("unmarshal images fail %s", err)
		return nil, 0, nil
	}
	total, _ := body.Int("TotalCount")
	return images, int(total), nil
}

func (self *SRegion) DeleteImage(imageId string) error {
	params := make(map[string]string)
	params["RegionId"] = self.RegionId
	params["ImageId"] = imageId
	params["Force"] = "true"

	_, err := self.ecsRequest("DeleteImage", params)
	if err != nil {
		log.Errorf("DeleteImage fail %s", err)
		return err
	}
	return nil
}
