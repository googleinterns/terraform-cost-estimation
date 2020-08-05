package jsdecode

import (
	"io/ioutil"
	"os"
	tfjson "github.com/hashicorp/terraform-json"
)

// ExtractPlanStruct extracts tfjson.Plan struct from file in provided path
// and returns the pointer on it if it is possible, otherwise returns error.
func ExtractPlanStruct (filePath string) (*tfjson.Plan, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	byteFile, _ := ioutil.ReadAll(f)

	var plan tfjson.Plan
	if err = plan.UnmarshalJSON(byteFile); err != nil {
		return nil, err
	}
	return &plan, nil
}
