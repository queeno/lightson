resource "google_project" "core" {
  name       = "core"
  project_id = var.project_id
  
  lifecycle {
    ignore_changes = [
      billing_account,
    ]
  }
}
