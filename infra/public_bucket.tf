resource "google_storage_bucket" "simon_public" {
  name = "simon-public"
  location = var.region

  versioning {
    enabled = true
  }
}

resource "google_storage_bucket_object" "cv_pdf" {
  name   = "simon-aquino-cv.pdf"
  source = "../assets/SimonAquinoCV.pdf"
  bucket = google_storage_bucket.simon_public.name
  cache_control = "no-store"
}

resource "google_storage_bucket_object" "cv_txt" {
  name   = "simon-aquino-cv.txt"
  source = "../assets/SimonAquinoCV.txt"
  bucket = google_storage_bucket.simon_public.name
  cache_control = "no-store"
}

resource "google_storage_object_access_control" "cv_pdf" {
  object = google_storage_bucket_object.cv_pdf.output_name
  bucket = google_storage_bucket.simon_public.name
  role   = "READER"
  entity = "allUsers"
}

resource "google_storage_object_access_control" "cv_txt" {
  object = google_storage_bucket_object.cv_txt.output_name
  bucket = google_storage_bucket.simon_public.name
  role   = "READER"
  entity = "allUsers"
}


