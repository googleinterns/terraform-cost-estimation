package main

import (
	"flag"
	"fmt"
	"os"
)

var out = flag.String("out", "stdout", `Write the cost estimation to a given file path. If set to 'stdout',
the output will be shown in the command line.`)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: go run main.go [OPTIONS] FILE\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), `  Outputs the cost estimation of Terraform resources from a JSON plan file.`)
		fmt.Fprintf(flag.CommandLine.Output(), "\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Fprintf(os.Stderr, "Error: No input file\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if len(flag.Args()) > 1 {
		fmt.Fprintf(os.Stderr, "Error: Too many argumets\n\n")
		flag.Usage()
		os.Exit(1)
	}

	fin, errFin := os.Open(flag.Args()[0])
	if errFin != nil {
		fmt.Fprintf(os.Stderr, "Error: "+errFin.Error()+"\n")
		os.Exit(1)
	}

	// TODO: use jsdecode package to get resources

	if errFin = fin.Close(); errFin != nil {
		fmt.Fprintf(os.Stderr, "Error: "+errFin.Error()+"\n")
		os.Exit(1)
	}

	// TODO: get pricing info

	// TODO: print output
}
