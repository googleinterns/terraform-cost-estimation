package classdetail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
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

// GetMachineDetails returns the number of cores and GBs of memory for a specific machine type.
func GetMachineDetails(machineType string) (coreNum int, memGB float64, err error) {
	if machineTypes == nil {
		machineTypes, err = getMachineTypes()
		if err != nil {
			return 0, 0, err

		}
	}

	d, ok := machineTypes[machineType]
	if !ok {
		return 0, 0, fmt.Errorf("machine type not supported")
	}
	return d.CoreNumber, d.MemoryGB, nil
}
