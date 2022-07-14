resource "google_storage_bucket" "simon_core_terraform" {
  name     = "simon-core-terraform"
  location = "us-east1"

  versioning {
    enabled = true
  }
}

resource "google_service_account" "terraform" {
  account_id   = "terraform"
  display_name = "Service Account for Terraform access"
}