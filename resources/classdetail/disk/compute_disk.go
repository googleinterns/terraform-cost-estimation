package disk

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type diskJSON struct {
	Name              string
	Region            string
	Zone              string
	DefaultDiskSizeGb string
	ValidDiskSize     string
}

// Disk holds information about disk size depending on type and zone/region.
type Disk struct {
	Type           string
	Region         string
	Zone           string
	DefaultSizeGiB int64
	MinSize        int64
	MaxSize        int64
}

func convertToDisk(d diskJSON) (*Disk, error) {
	final := &Disk{Type: d.Name, Region: d.Region, Zone: d.Zone}

	def, err := strconv.ParseInt(d.DefaultDiskSizeGb, 10, 64)
	if err != nil {
		return nil, err
	}
	final.DefaultSizeGiB = def

	// Separate min and max strings and remove trailing "GB" before conversion.
	i := strings.LastIndex(d.ValidDiskSize, "-")
	if i < 0 {
		return nil, fmt.Errorf("invalid disk size interval format")
	}

	min, err := strconv.ParseInt(d.ValidDiskSize[:i-2], 10, 64)
	if err != nil {
		return nil, err
	}
	final.MinSize = min

	max, err := strconv.ParseInt(d.ValidDiskSize[i+1:len(d.ValidDiskSize)-2], 10, 64)
	if err != nil {
		return nil, err
	}
	final.MaxSize = max

	return final, nil
}

// ReadDiskInfo reads the JSON file with disk information.
func ReadDiskInfo() (map[string]map[string]*Disk, error) {
	var diskTypes map[string]map[string]*Disk

	// Get json file path relative to this directory.
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/compute_disk_types.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	var jsonMap []diskJSON
	json.Unmarshal(data, &jsonMap)

	diskTypes = map[string]map[string]*Disk{}
	for _, d := range jsonMap {
		d2, err := convertToDisk(d)
		if err != nil {
			return nil, err
		}
		if diskTypes[d.Name] == nil {
			diskTypes[d.Name] = map[string]*Disk{}
		}

		if d.Zone != "" {
			diskTypes[d.Name][d.Zone] = d2
		} else {
			diskTypes[d.Name][d.Region] = d2
		}
	}

	return diskTypes, nil
}

// Details returns default, minimum and maximum size (in GiB) of a disk type running in the specific zone or region.
// If the combination of disk type and location is invalid, an error is returned.
func Details(diskTypes map[string]map[string]*Disk, diskType, zone, region string) (int64, int64, int64, error) {
	if diskTypes == nil {
		return 0, 0, 0, fmt.Errorf("disk details are not initialized")
	}

	d1, ok := diskTypes[diskType]
	if !ok {
		return 0, 0, 0, fmt.Errorf("invalid disk type '" + diskType + "'")
	}

	d2, ok := d1[zone]
	if !ok {
		d2, ok = d1[region]
		if !ok {
			return 0, 0, 0, fmt.Errorf("no disk type running in '" + region + "'")
		}
	}
	return d2.DefaultSizeGiB, d2.MinSize, d2.MaxSize, nil
}
