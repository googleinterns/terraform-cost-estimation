package instance

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	memconv "github.com/googleinterns/terraform-cost-estimation/memconverter"
)

// ComputeInstanceInfo holds information about compute instance types.
type ComputeInstanceInfo struct {
	CoreNumber int
	MemoryGiB  float64
}

var sharedCoreFractional = map[string]float64{
	"e2-micro":  0.125,
	"e2-small":  0.25,
	"e2-medium": 0.5,
	"f1-micro":  0.2,
	"g1-small":  0.5,
}

// ReadMachineTypes reads JSON file with compute instance type information.
func ReadMachineTypes() (map[string]ComputeInstanceInfo, error) {
	// Get path to JSON  file relative to this directory.
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/machine_types.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]ComputeInstanceInfo
	json.Unmarshal(data, &jsonMap)
	return jsonMap, nil
}

// getCustomMachineDetails looks for a cutom machine type in the format [machine_type-]custom-<core_num>-<mem_mib>[-ext] and extracts <core_num> and <mem_mib>.
func getCustomMachineDetails(machineType string) (coreNum int, memGiB float64, err error) {
	// Remove '-ext' if needed.
	if strings.HasSuffix(machineType, "-ext") {
		i := strings.LastIndex(machineType, "-")
		machineType = machineType[:i]
	}

	// Look for <mem_mib> string.
	i := strings.LastIndex(machineType, "-")
	if i < 0 {
		return 0, 0, fmt.Errorf("invalid custom machine type format")
	}
	memStr := machineType[i+1:]
	machineType = machineType[:i]

	// Look for <core_num> string.
	i = strings.LastIndex(machineType, "-")
	if i < 0 {
		return 0, 0, fmt.Errorf("invalid custom machine type format")
	}
	coresStr := machineType[i+1:]

	// Convert to numbers and GiB memory unit.
	coreNum, err = strconv.Atoi(coresStr)
	if err != nil {
		return 0, 0, err
	}

	memMiB, err := strconv.Atoi(memStr)
	if err != nil {
		return 0, 0, err
	}

	memGiB, err = memconv.Convert("mib", float64(memMiB), "gib")
	if err != nil {
		return 0, 0, err
	}
	return
}

// GetMachineDetails returns the number of cores and GBs of memory for a specific machine type.
func GetMachineDetails(machineTypes map[string]ComputeInstanceInfo, machineType string) (coreNum int, memGiB float64, err error) {
	if machineTypes == nil {
		return 0, 0, fmt.Errorf("machine type information is not initialized")
	}

	if strings.Contains(machineType, "custom") {
		return getCustomMachineDetails(machineType)
	}

	d, ok := machineTypes[machineType]
	if !ok {
		return 0, 0, fmt.Errorf("machine type not supported")
	}
	return d.CoreNumber, d.MemoryGiB, nil
}

// GetMachineFractionalCore returns fractional vCPU of the machine type.
// For non-shared-core machines, the return value is 1.
func GetMachineFractionalCore(machineType string) float64 {
	if d, ok := sharedCoreFractional[machineType]; ok {
		return d
	}
	return 1
}
