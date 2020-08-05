package jsdecode

import (
	"fmt"
	"io/ioutil"
	"os"
	tfjson "github.com/hashicorp/terraform-json"
)

// Function extracts tfjson.Plan struct from file in provided path and
// returns the pointer on it if it is possible, otherwise return error.
func ExtractPlanStruct (filePath string) (*tfjson.Plan, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	byteFile, _ := ioutil.ReadAll(f)

	var plan tfjson.Plan
	err = plan.UnmarshalJSON(byteFile)
	if err != nil {
		return &plan, err
	}
	return &plan, nil
}
