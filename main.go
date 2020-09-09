package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io"
	"github.com/googleinterns/terraform-cost-estimation/jsdecode"
)

var (
	output = flag.String("output", "stdout", `Write the cost estimations to the given paths.
If set to 'stdout', all the outputs will be shown in the command line.
Multiple output file names must be delimited by ','.
Mixed file names and stdout values are allowed.`)
	format = flag.String("format", "txt", `Write the pricing information in the specified format.
Can be set to: txt, json, html.`)
)

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}

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

	outputFileNames := strings.Split(*output, ",")
	if *output != "stdout" {
		if len(outputFileNames) != len(flag.Args()) {
			fmt.Fprintf(os.Stderr, "Error: Input and output files number differ.\n\n")
			flag.Usage()
			os.Exit(1)
		}
	}

	computeEngineCatalog, catalogErr := billing.NewComputeEngineCatalog(context.Background())
	if catalogErr != nil {
		fmt.Fprintf(os.Stderr, "Error: "+catalogErr.Error()+"\n\n")
		os.Exit(2)
	}

	for i := range flag.Args() {
		plan, err := io.GetPlan(flag.Arg(i))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: "+err.Error()+"\n\n")
			os.Exit(3)
		}

		resources := jsdecode.GetResources(plan)

		fout, err := io.GetOutputWriter(outputFileNames[minInt(i, len(outputFileNames)-1)])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: "+err.Error()+"\n\n")
		}

		fout.Write([]byte(fmt.Sprintf("Pricing information for %s:\n\n", flag.Arg(i))))
		summary := ""

		for _, r := range resources {
			if err = r.CompletePricingInfo(computeEngineCatalog); err != nil {
				fmt.Fprintf(os.Stderr, "Error: "+err.Error()+"\n\n")
				continue
			}
			r.PrintPricingInfo(fout)
			summary += r.GetSummary() + "\n"
		}

		fout.Write([]byte("\n\nSummary:\n\n" + summary))

		if err = io.FinishOutput(fout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: "+err.Error()+"\n\n")
			os.Exit(3)
		}
	}
}
