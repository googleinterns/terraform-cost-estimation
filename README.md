# Cost Estimation for GCP infrastructure deployed by Terraform
[![Build Status](https://travis-ci.org/googleinterns/terraform-cost-estimation.svg?branch=master)](https://travis-ci.org/googleinterns/terraform-cost-estimation)

The terraform-cost-estimation project will surface a before/after cost
estimation for GCP infrastructure deployed by Terraform.

**This is not an officially supported Google product.**

Display before/after cost estimation of resources from Terraform plan files in JSON format.

**Resources supported:**
- **google_compute_instance**
- **google_compute_disk**

## Usage
In the command line, run: __go run main.go [OPTIONS] FILES__

## Options:
- **format**
	- Write the pricing information in the specified format.
    - Can be set to: txt, json, html.
    - If omitted, if defaults to 'txt'.

- **output**
	- Write the cost estimations to the given paths.
    - If set to 'stdout', all the outputs will be shown in the command line.
    - Multiple output file names must be delimited by ','.
    - Mixed file names and stdout values are allowed.

**Examples:**
- **go run main.go input.json**
- **go run main.go -output=json input.json**
- **go run main.go -format=html -output=out1.html,out2.html input1.json input2.json**