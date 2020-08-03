package main

import (
    "fmt"
    "os"
    "encoding/json"
    "io/ioutil"
)
type ComputeInstanceInfo struct {
    Name string
    MachineType string
    Zone string

    IsNil bool
}

type ComputeInstance struct {
    BeforeState ComputeInstanceInfo
    AfterState ComputeInstanceInfo
}

//extract the info about compute instance from go interface
//we call it for before and after json for every change in resource_changes
func ExtractInstanceInfo (change interface{}) ComputeInstanceInfo {
    var computeInstance ComputeInstanceInfoa
    if change != nil {
        change := change.(map[string]interface{})
        computeInstance.IsNil = false
        computeInstance.Name = change["name"].(string)
        computeInstance.MachianeType = change["machine_type"].(string)
        computeInstance.Zone = change["zone"].(string)
    } else {
        computeInstance.IsNil = true
    }
    return computeInstance
}
func main() {
 	fmt.Println("JSON file name:", os.Args[1])
	jsonFileName := os.Args[1]
    jsonFile, err := os.Open(jsonFileName)

    if err != nil {
	// Handle error.
      	fmt.Println(err)
    }
    fmt.Println("Successfully opened file.")
    defer jsonFile.Close()
    
    byteValue, _ := ioutil.ReadAll(jsonFile)
    var result map[string]interface{}
    json.Unmarshal([]byte(byteValue), &result)

    //list of the resources changes
    resourceChanges := result["resource_changes"].([]interface{})

    //list of changed compute instances
    var ChangedComputeInstances []ComputeInstance


    var computeInstance ComputeInstance
    // var beforeInfo ComputeInstanceInfo
    // var afterInfo ComputeInstanceInfo
    for _, c := range resourceChanges {
        //cast c to map struct
        cmap := c.(map[string]interface{})
        change := cmap["change"].(map[string]interface{})
    
        // beforeInfo = ExtractInstanceInfo(change["before"])
        // afterInfo = ExtractInstanceInfo(change["after"])
        // fmt.Println(beforeInfo.Name)
        // fmt.Println(beforeInfo.MachineType)
        // fmt.Println(beforeInfo.Zone)
        // fmt.Println(afterInfo.Name)

        computeInstance.BeforeState = ExtractInstanceInfo(change["before"])
        computeInstance.AfterState = ExtractInstanceInfo(change["after"])
        fmt.Println(computeInstance.BeforeState.Name)
        fmt.Println(computeInstance.BeforeState.MachineType)
        fmt.Println(computeInstance.AfterState.Name)
        ChangedComputeInstances = append(ChangedComputeInstances, computeInstance)
    }
    fmt.Println(ChangedComputeInstances)
}
