terraform {
  backend "gcs" {
    bucket = "simon-core-terraform"
    prefix = "terraform-core"
  }
}