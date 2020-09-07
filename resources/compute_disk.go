package resources

import (
	"fmt"
	"strings"

	dsk "github.com/googleinterns/terraform-cost-estimation/resources/classdetail/disk"
	img "github.com/googleinterns/terraform-cost-estimation/resources/classdetail/image"
)

// ComputeDisk holds information about the compute disk resource type.
type ComputeDisk struct {
	Name        string
	ID          string
	Type        string
	Zone        string
	Region      string
	Image       string
	Snapshot    string
	SizeGiB     int64
	UnitPricing PricingInfo
}

// NewComputeDisk builds a compute disk with the specified fields and fills the other resource details.
// Image, snapshot and size parameters are considered null fields when "" or <= 0.
// Returns a pointer to a ComputeInstance structure, or nil and error upon failure.
// Currently not supported: snapshots.
func NewComputeDisk(name, id, diskType, zone, image, snapshot string, size int64) (*ComputeDisk, error) {
	disk := &ComputeDisk{Name: name, ID: id, Type: diskType, Zone: zone}

	i := strings.LastIndex(zone, "-")
	if i < 0 {
		return nil, fmt.Errorf("invalid zone format")
	}
	disk.Region = zone[:i]

	def, min, max, err := dsk.Details(diskType, zone, disk.Region)
	if err != nil {
		return nil, err
	}

	switch {
	case image == "":
		if size <= 0 {
			disk.SizeGiB = def
		} else {
			disk.SizeGiB = size
		}

	case size <= 0:
		s, err := img.GetImageDiskSize(image)
		if err != nil {
			return nil, err
		}
		disk.SizeGiB = s

	default:
		s, err := img.GetImageDiskSize(image)
		if err != nil {
			return nil, err
		}
		if size < s {
			return nil, fmt.Errorf("size should at least be the size of the specified image")
		}
		disk.SizeGiB = size
	}

	if disk.SizeGiB < min || disk.SizeGiB > max {
		return nil, fmt.Errorf("size is not in the valid range")
	}

	return disk, nil
}
