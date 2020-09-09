package io

import (
	"os"

	"github.com/googleinterns/terraform-cost-estimation/jsdecode"
	tfjson "github.com/hashicorp/terraform-json"
)

// GetPlan receives the input file name to extract the plan structure or return an error.
// The file is also closed if it was successfully opened.
func GetPlan(inputName string) (*tfjson.Plan, error) {
	fin, err := os.Open(inputName)
	if err != nil {
		return nil, err
	}

	defer fin.Close()

	plan, err := jsdecode.ExtractPlanStruct(fin)
	if err != nil {
		return nil, err
	}

	return plan, nil
}
