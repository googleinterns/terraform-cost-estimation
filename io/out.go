package io

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/googleinterns/terraform-cost-estimation/io/js"
	"github.com/googleinterns/terraform-cost-estimation/io/web"
	"github.com/googleinterns/terraform-cost-estimation/resources"
	"github.com/jedib0t/go-pretty/v6/table"
	"html/template"
	"log"
)

// GetOutputWriter returns the output os.File (stdout/file) for a given output path or an error.
func GetOutputWriter(outputPath string) (*os.File, error) {
	if outputPath == "stdout" {
		return os.Stdout, nil
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// FinishOutput closes the file when the output is done and returns an error where is the case.
// If the output file is Stdout, thenit is not closed and 2 newlines are printed so that different plan file outputs can be separated.
func FinishOutput(outputFile *os.File) error {
	if outputFile != os.Stdout {
		return outputFile.Close()
	}
	fmt.Println("\n-----------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("\n\n\n")
	return nil
}

// GenerateWebPage generates a html output with the pricing information of the specified resources.
func GenerateWebPage(f *os.File, res []resources.ResourceState) error {
	// Get path of template relative to this file.
	_, callerFile, _, _ := runtime.Caller(0)
	t, err := template.ParseFiles(filepath.Dir(callerFile) + "/web/web_template.gohtml")
	if err != nil {
		return err
	}

	if err = t.Execute(f, mapToWebTables(res)); err != nil {
		return err
	}

	return nil
}

func mapToWebTables(res []resources.ResourceState) (t []*web.PricingTypeTables) {
	for i, r := range res {
		t = append(t, r.GetWebTables(i))
	}
	return
}

// RenderJson returns the string with json output struct for all resources.
func RenderJson(states []resources.ResourceState) (string, error) {
	out := js.JsonOutput{}
	out.Delta = getTotalDelta(states)
	out.PricingUnit = "USD/hour"
	for _, state := range states {
		s, err := state.ToStateOut()
		if err == nil || s != nil {
			s.AddToJSONTableList(&out)
		}
	}
	jsonString, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(jsonString), err
}

// GenerateJsonOut generates a json file with the pricing information of the specified resources.
func GenerateJsonOut(f *os.File, res []resources.ResourceState) error {
	jsonString, err := RenderJson(res)
	if err != nil {
		return nil
	}
	if _, err = io.WriteString(f, jsonString); err != nil {
		return err
	}
	return nil
}

// GetSummaryTable returns the table with brief cost changes info about all resources (Compute Instances and Compute Disks).
func GetSummaryTable(states []resources.ResourceState) *table.Table {
	t := &table.Table{}
	autoMerge := table.RowConfig{AutoMerge: true}

	dTotal := getTotalDelta(states)
	t.SetTitle(fmt.Sprintf("The total cost change for all Resources is %.6f USD/hour.", dTotal))
	h := "Pricing Information\n(USD/h)"
	t.AppendRow(table.Row{h, h, h, h, h}, autoMerge)
	t.AppendRow(table.Row{"Name", "ID", "Type", "Action", "Delta"})
	for _, s := range states {
		if row, err := s.GetSummaryRow(); err == nil {
			t.AppendRow(row)
		} else {
			log.Printf("Error: %v", err)
		}
	}
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	return t
}

// OutputPricing writes pricing information about each resource and summary.
func OutputPricing(states []resources.ResourceState, f *os.File) {
	f.Write([]byte(GetSummaryTable(states).Render() + "\n\n"))
	f.Write([]byte("\n List of all Resources:\n\n"))
	for _, s := range states {
		if s != nil {
			t, err := s.ToTable()
			if err == nil {
				f.Write([]byte(t.Render() + "\n\n\n"))
			} else {
				log.Printf("Error: %v", err)
			}
		}
	}
}

// getTotalDelta returns the cost change of all resources.
func getTotalDelta(states []resources.ResourceState) float64 {
	var t float64
	for _, s := range states {
		t += s.GetDelta()
	}
	return t
}
