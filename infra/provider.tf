provider "google" {
  project     = var.project_id
  region      = var.region
  credentials = file("${var.google_creds}")
}

provider "google-beta" {
  project     = var.project_id
  region      = var.region
  credentials = file("${var.google_creds}")
}
