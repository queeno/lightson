provider "google" {
  project     = var.project_id
  region      = var.region
  credentials = "account.json"
}

provider "google-beta" {
  project     = var.project_id
  region      = var.region
  credentials = "account.json"
}