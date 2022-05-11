provider "poetry" {
}

terraform {
  required_providers {
    poetry = {
      source  = "max-frank/poetry"
      version = "~> 1.0"
    }
  }
}
