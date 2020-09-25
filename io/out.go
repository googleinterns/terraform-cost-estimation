package io

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"html/template"

	"github.com/googleinterns/terraform-cost-estimation/io/web"
	"github.com/googleinterns/terraform-cost-estimation/resources"
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
