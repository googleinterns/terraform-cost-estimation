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

type disk struct {
	Type           string
	Region         string
	Zone           string
	DefaultSizeGiB int64
	MinSize        int64
	MaxSize        int64
}

// Disks are stored first by type, then by zone/region.
var diskTypes map[string]map[string]*disk

func convertToDisk(d diskJSON) (*disk, error) {
	final := &disk{Type: d.Name, Region: d.Region, Zone: d.Zone}

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

func readDiskInfo() error {
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/compute_disk_types.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return err
	}

	var jsonMap []diskJSON
	json.Unmarshal(data, &jsonMap)

	diskTypes = map[string]map[string]*disk{}
	for _, d := range jsonMap {
		d2, err := convertToDisk(d)
		if err != nil {
			return err
		}
		if diskTypes[d.Name] == nil {
			diskTypes[d.Name] = map[string]*disk{}
		}

		if d.Zone != "" {
			diskTypes[d.Name][d.Zone] = d2
		} else {
			diskTypes[d.Name][d.Region] = d2
		}
	}

	return nil
}

// Details returns default, minimum and maximum size (in GiB) of a disk type running in the specific zone or region.
// If the combination of disk type and location is invalid, an error is returned.
func Details(diskType, zone, region string) (int64, int64, int64, error) {
	if diskTypes == nil {
		if err := readDiskInfo(); err != nil {
			return 0, 0, 0, err
		}
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
