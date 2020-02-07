resource "google_storage_bucket" "simon_public" {
  name = "simon-public"
  location = var.region

  versioning {
    enabled = true
  }
}

resource "google_storage_bucket_object" "cv" {
  name   = "simon-aquino-cv.pdf"
  source = "../assets/SimonAquinoCV.pdf"
  bucket = google_storage_bucket.simon_public.name
}

resource "google_storage_object_access_control" "cv" {
  object = google_storage_bucket_object.cv.output_name
  bucket = google_storage_bucket.simon_public.name
  role   = "READER"
  entity = "allUsers"
}

