variable "region" {}

module "m1" {
  source = "./module_b"
  region = var.region
  zone   = "a"
}

module "m2" {
  source = "./module_b"
  region = var.region
  zone   = "b"
}
