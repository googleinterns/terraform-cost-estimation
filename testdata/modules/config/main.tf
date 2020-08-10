module "m1" {
  source = "./module_a"
  region = "us-central1"
}

module "m2" {
  source = "./module_a"
  region = "us-east1"
}

module "m3" {
  source = "./module_a"
  region = "us-west1"
}

resource "null_resource" "abc" {
}
