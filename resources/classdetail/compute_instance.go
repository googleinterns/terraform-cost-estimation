package classdetail

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

type computeInstance struct {
	CoreNumber int
	MemoryGB   float64
}

var machineTypes map[string]computeInstance

func getMachineTypes() (map[string]computeInstance, error) {
	_, callerFile, _, _ := runtime.Caller(0)
	inputPath := filepath.Dir(callerFile) + "/machine_types.json"

	data, err := ioutil.ReadFile(inputPath)
	if err != nil {
		return nil, err
	}

	var jsonMap map[string]computeInstance
	json.Unmarshal(data, &jsonMap)
	return jsonMap, nil
}

func getCustomMachineDetails(machineType string) (coreNum int, memGiB float64, err error) {
	if strings.HasSuffix(machineType, "-ext") {
		i := strings.LastIndex(machineType, "-")
		machineType = machineType[:i]
	}

	i := strings.LastIndex(machineType, "-")
	if i < 0 {
		return 0, 0, fmt.Errorf("invalid custom machine type format")
	}
	memStr := machineType[i+1:]
	machineType = machineType[:i]

	i = strings.LastIndex(machineType, "-")
	if i < 0 {
		return 0, 0, fmt.Errorf("invalid custom machine type format")
	}
	coresStr := machineType[i+1:]

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
func GetMachineDetails(machineType string) (coreNum int, memGB float64, err error) {
	if machineTypes == nil {
		machineTypes, err = getMachineTypes()
		if err != nil {
			return 0, 0, err

		}
	}

	if strings.Contains(machineType, "custom") {
		return getCustomMachineDetails(machineType)
	}

	d, ok := machineTypes[machineType]
	if !ok {
		return 0, 0, fmt.Errorf("machine type not supported")
	}
	return d.CoreNumber, d.MemoryGB, nil
}
