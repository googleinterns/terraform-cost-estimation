package io

import (
	"fmt"
	"os"
)

// GetOutputWriter returns the output location (stdout/file) for a given input file or an error.
func GetOutputWriter(outputPath string) (*os.File, error) {
	if outputPath == "stdout" {
		fmt.Println("-----------------------------------------------------------------------------------------------------------------------------")
		return os.Stdout, nil
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// FinishOutput closes the file when the output is done and returns an error where is the case.
// If the output file is Stdout, thenit is not closed and 2 newlines are printed.
func FinishOutput(outputFile *os.File) error {
	if outputFile != os.Stdout {
		return outputFile.Close()
	}
	fmt.Println("\n-----------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("\n\n\n")
	return nil
}
