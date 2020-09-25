package classdetail

import (
	"github.com/googleinterns/terraform-cost-estimation/resources/classdetail/disk"
	"github.com/googleinterns/terraform-cost-estimation/resources/classdetail/image"
	"github.com/googleinterns/terraform-cost-estimation/resources/classdetail/instance"
)

// ResourceDetail holds information about resource details.
// Disk type information is stored first by disk type, then by zone/region.
type ResourceDetail struct {
	diskInfo     map[string]map[string]*disk.Disk
	imageInfo    *image.ImageInfo
	instanceInfo map[string]instance.ComputeInstanceInfo
}

// NewResourceDetail builds a ResourceDetail object.
func NewResourceDetail() (*ResourceDetail, error) {
	rd := &ResourceDetail{}

	// Initialize disk information.
	d, err := disk.ReadDiskInfo()
	if err != nil {
		return nil, err
	}
	rd.diskInfo = d

	// Initialize image information.
	i, err := image.ReadComputeImagesInfo()
	if err != nil {
		return nil, err
	}
	rd.imageInfo = i

	// Initialize compute instance machine type information.
	m, err := instance.ReadMachineTypes()
	if err != nil {
		return nil, err
	}
	rd.instanceInfo = m

	return rd, nil
}

// DiskDetails returns default, minimum and maximum size (in GiB) of a disk type running in the specific zone or region.
func (rd *ResourceDetail) DiskDetails(diskType, zone, region string) (int64, int64, int64, error) {
	return disk.Details(rd.diskInfo, diskType, zone, region)
}

// ImageSize returns the size of a compute image.
func (rd *ResourceDetail) ImageSize(img string) (int64, error) {
	return image.GetImageDiskSize(rd.imageInfo, img)
}

// MachineDetails returns the number of cores and amount of memory (in GiB) of a compute instance type.
func (rd *ResourceDetail) MachineDetails(machineType string) (coreNum int, memGiB float64, err error) {
	return instance.GetMachineDetails(rd.instanceInfo, machineType)
}

// MachineFractionalCore returns the fractional core value of a compute instance type.
func (rd *ResourceDetail) MachineFractionalCore(machineType string) float64 {
	return instance.GetMachineFractionalCore(machineType)
}
