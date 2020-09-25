package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/googleinterns/terraform-cost-estimation/billing"
	"github.com/googleinterns/terraform-cost-estimation/io"
	"github.com/googleinterns/terraform-cost-estimation/jsdecode"
	res "github.com/googleinterns/terraform-cost-estimation/resources"
	cd "github.com/googleinterns/terraform-cost-estimation/resources/classdetail"
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
		fmt.Fprintf(flag.CommandLine.Output(), "Outputs the cost estimation of Terraform resources from a JSON plan file.")
		fmt.Fprintf(flag.CommandLine.Output(), "\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("Error: No input file.")
	}

	outputs := strings.Split(*output, ",")
	if *output != "stdout" {
		if len(outputs) != len(flag.Args()) {
			log.Fatal("Error: Input and output files number differ.")
		}
	}

	catalog, err := billing.NewComputeEngineCatalog(context.Background())
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	classDetails, err := cd.NewResourceDetail()
	if err != nil {
		log.Fatalf("Error: %+v", err)
	}

	for i, inputName := range flag.Args() {
		plan, err := io.GetPlan(inputName)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		resources := jsdecode.GetResources(classDetails, plan)

		finalResources := []res.ResourceState{}
		for _, r := range resources {
			if err = r.CompletePricingInfo(catalog); err != nil {
				log.Printf("In file %s got error: %v", inputName, err)
				continue
			}
			finalResources = append(finalResources, r)
		}

		outputName := outputs[minInt(i, len(outputs)-1)]
		fout, err := io.GetOutputWriter(outputName)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		switch {
		case *format == "json":
			if err = res.GenerateJsonOut(fout, finalResources); err != nil {
				log.Printf("Error: %v", err)
			}
		case *format == "html":
			if err = io.GenerateWebPage(fout, finalResources); err != nil {
				log.Printf("Error: %v", err)
			}
		case *format == "txt":
			res.OutputPricing(finalResources, fout)
		default:
		}

		if err = io.FinishOutput(fout); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
