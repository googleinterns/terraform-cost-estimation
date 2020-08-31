package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/jsdecode"
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

	plan, err := jsdecode.ExtractPlanStruct(fin)
	if err != nil {
		fmt.Fprint(os.Stderr, "Error:"+err.Error()+"\n")
		os.Exit(2)
	}

	if errFin = fin.Close(); errFin != nil {
		fmt.Fprintf(os.Stderr, "Error: "+errFin.Error()+"\n")
		os.Exit(1)
	}

	resources := jsdecode.GetResources(plan)

	computeEngineCatalog, catalogErr := billing.NewComputeEngineCatalog(context.Background())
	if catalogErr != nil {
		fmt.Fprintf(os.Stderr, "Error: "+catalogErr.Error())
		os.Exit(4)
	}

	var fout *os.File
	var errFout error
	if *out == "stdout" {
		fout = os.Stdout
	} else {
		fout, errFout = os.Open(*out)
		if errFout != nil {
			fmt.Fprintf(os.Stderr, "Error: "+errFout.Error()+"\n")
			os.Exit(3)
		}
	}

	for _, r := range resources {
		r.CompletePricingInfo(computeEngineCatalog)
		r.PrintPricingInfo(fout)
	}

	if errFout = fout.Close(); errFout != nil {
		fmt.Fprintf(os.Stderr, "Error: "+errFout.Error()+"\n")
		os.Exit(1)
	}

	// TODO: print output
}
