project = "gcp-artifact-registry-test"

app "example" {
  build {
    use "docker" {
      buildkit           = false
      disable_entrypoint = false
    }

    registry {
      use "gcp-artifact-registry" {
        project       = "cnfi-305306"
        location      = "us"
        repository_id = "test"
      }
    }
  }

  deploy {
    use "docker" {

    }
  }


}

