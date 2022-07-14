#module "roh_scout" {
#  source                         = "terraform-google-modules/scheduled-function/google"
#  version                        = "2.5.0"
#  project_id                     = google_project.core.project_id
#  job_name                       = "roh-scout"
#  job_schedule                   = "*/5 * * * *"
#  function_entry_point           = "Run"
#  function_source_directory      = format("%s/roh-scout", path.module)
#  function_name                  = "roh-scout"
#  region                         = "us-east1"
#  function_runtime               = "go116"
#  topic_name                     = "roh-scout"
#  function_service_account_email = google_service_account.roh_scout.email
#  function_environment_variables = {
#    BUCKET             = google_storage_bucket.roh_scout.id
#    TICKETS_WANTED     = 2
#    PERFORMANCE_IDS    = "48856,48857"
#    TELEGRAM_CHAT_ID   = 802003234
#    SEATS              = "18 Stalls Circle Standing"
#    TELEGRAM_SECRET_ID = google_secret_manager_secret.telegram_token.id
#  }
#}
#
#resource "google_service_account" "roh_scout" {
#  account_id   = google_project.core.project_id
#  display_name = "roh-scout"
#}
#
#resource "google_storage_bucket" "roh_scout" {
#  name                        = "roh-scout"
#  location                    = "us-east1"
#  force_destroy               = true
#  uniform_bucket_level_access = true
#}
#
#resource "google_storage_bucket_iam_member" "member" {
#  bucket = google_storage_bucket.roh_scout.name
#  role   = "roles/storage.admin"
#  member = "serviceAccount:${google_service_account.roh_scout.email}"
#}
#
#resource "google_secret_manager_secret" "telegram_token" {
#  secret_id = "roh-scout-telegram-token"
#  replication {
#    user_managed {
#      replicas {
#        location = "us-east1"
#      }
#    }
#  }
#}
#
#resource "google_secret_manager_secret_iam_member" "member" {
#  secret_id = google_secret_manager_secret.telegram_token.id
#  role      = "roles/secretmanager.secretAccessor"
#  member    = "serviceAccount:${google_service_account.roh_scout.email}"
#}
#