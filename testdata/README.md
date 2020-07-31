# Test fixtures

This directory contains test fixtures.

The Terraform plan files have been created by running `terraform plan
-out=tfplan` on a given Terraform configuration and then converting the binary
`tfplan` file into JSON by running `terraform show -json tfplan | jq`.
