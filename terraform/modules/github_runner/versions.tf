terraform {
  required_version = ">= 1.5"
  
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.9"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.9"
    }
  }
}
