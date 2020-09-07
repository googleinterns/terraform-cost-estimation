package io

import (
	"fmt"
	"os"

	"github.com/googleinterns/terraform-cost-estimation/resources"
)

const (
	webBeginning = `
<!DOCTYPE html>
<html lang="en">
	<link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.5.2/css/bootstrap.min.css" integrity="sha384-JcKb8q3iqJ61gNV9KGb8thSsNjpSL0n8PARn9HuZOnIxN0hoP+VmmDGMN5t9UJ0Z" crossorigin="anonymous">
	<script src="https://code.jquery.com/jquery-3.2.1.slim.min.js" integrity="sha384-KJ3o2DKtIkvYIK3UENzmM7KCkRr/rE9/Qpg6aAZGJwFDMVNA/GpGFF93hXpG5KkN" crossorigin="anonymous"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.12.9/umd/popper.min.js" integrity="sha384-ApNbgh9B+Y1QKtv3Rn7W3mgPxhU9K/ScQsAP7hUibX39j7fakFPskvXusvfa0b4Q" crossorigin="anonymous"></script>
	<script src="https://maxcdn.bootstrapcdn.com/bootstrap/4.0.0/js/bootstrap.min.js" integrity="sha384-JZR6Spejh4U02d8jOt6vLEHfe/JQGiRRSQQxSfFWpi1MquVdAyjUar5+76PVCmYl" crossorigin="anonymous"></script>
    <style>
        table {
            width: 50%;
            border-collapse: collapse;
        }
        .div-table {
            width: 70%;
            margin-left: 15%;
            margin-top: 3%;
        }
        .label tr td label {
            display: block;
        }
        [data-toggle="toggle"] {
            display: none;
        }
        .hidden {
            display:none;
        }
        .show_div {
            display:block;
        }
        td, th {
            text-align:center;
        }
    </style>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document</title>
</head>
<body>
    <div class="navbar navbar-dark bg-dark" >
        <div class="dropdown show">
            <a class="btn btn-secondary dropdown-toggle" href="#" role="button" id="dropdownMenuLink" data-toggle="dropdown" aria-haspopup="true" aria-expanded="false">
              Pricing
            </a>
            <div class="dropdown-menu" aria-labelledby="dropdownMenuLink">
              <a class="dropdown-item" href="#" onclick="toggler('hourly_tables');">Hourly</a>
              <a class="dropdown-item" href="#" onclick="toggler('monthly_tables');">Monthly</a>
              <a class="dropdown-item" href="#" onclick="toggler('yearly_tables');">Yearly</a>
            </div>
        </div>
    </div>
	`

	webEnding = `
</body>
    <script>
        var divId_pre = "hourly_tables";
        function toggler(divId) {
            $("#" + divId).toggle();
            $("#" + divId_pre).toggle();
            divId_pre = divId;
        }
    </script>
    <script src="node_modules/jquery/dist/jquery.slim.min.js"></script>
    <script src="node_modules/popper.js/dist/umd/popper.min.js"></script>
    <script src="node_modules/bootstrap/dist/js/bootstrap.min.js"></script>
</html>
	`
)

// GetOutputWriter returns the output os.File (stdout/file) for a given output path or an error.
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
	if outputFile == nil {
		return nil
	}
	if outputFile != os.Stdout {
		return outputFile.Close()
	}
	fmt.Println("\n-----------------------------------------------------------------------------------------------------------------------------")
	fmt.Printf("\n\n\n")
	return nil
}

// GenerateWebPage generates a webpage file with the pricing information of the specified resources.
func GenerateWebPage(outputPath string, res []*resources.ComputeInstanceState) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(webBeginning)
	if err != nil {
		return err
	}

	divs := getTableDivs(res)
	_, err = f.WriteString(divs)
	if err != nil {
		return err
	}

	_, err = f.WriteString(webEnding)
	if err != nil {
		return err
	}

	return nil
}

func getTableDivs(res []*resources.ComputeInstanceState) (divs string) {
	ht := ""
	mt := ""
	yt := ""
	for i, r := range res {
		h, m, y := r.GetWebTables(i)
		ht += "\n" + h
		mt += "\n" + m
		yt += "\n" + y
	}

	beginning := `
    <div class="div-table %s" id="%s">
    `
	ending := `
    </div>
    `

	hourly := fmt.Sprintf(beginning, "show_div", "hourly_tables") + "\n" + ht + "\n" + ending
	monthly := fmt.Sprintf(beginning, "hidden", "monthly_tables") + "\n" + mt + "\n" + ending
	yearly := fmt.Sprintf(beginning, "hidden", "yearly_tables") + "\n" + yt + "\n" + ending

	return hourly + "\n" + monthly + "\n" + yearly + "\n"
}
