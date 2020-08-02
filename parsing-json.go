package main

import (
    "fmt"
    "os"
    "encoding/json"
    "io/ioutil"
)

func main() {
	fmt.Println("It works! =)")
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
    resourceChanges := result["resource_changes"].([]interface{})

    for _, c := range resourceChanges {
        cmap := c.(map[string]interface{})
        change := cmap["change"].(map[string]interface{})
        before := change["before"]
        if before != nil {
            before := before.(map[string]interface{})
            fmt.Println(before)
        } else {
            fmt.Println("before = nil")
        }
    }
}
