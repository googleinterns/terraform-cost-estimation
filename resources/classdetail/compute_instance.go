package classdetail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type computeInstance struct {
	CoreNumber int
	MemoryGB   float64
}

var machineTypes = getMachineTypes()

func getMachineTypes() map[string]computeInstance {
	f, _ := os.Open("machine_types.json")
	defer f.Close()

	data, _ := ioutil.ReadAll(f)

	var jsonMap map[string]computeInstance
	json.Unmarshal(data, &jsonMap)
	return jsonMap
}

// GetMachineDetails returns the number of cores and GBs of memory for a specific machine type.
func GetMachineDetails(machineType string) (coreNum int, memGB float64, err error) {
	d, ok := machineTypes[machineType]
	if !ok {
		return 0, 0, fmt.Errorf("machine type not supported")
	}
	return d.CoreNumber, d.MemoryGB, nil
}
