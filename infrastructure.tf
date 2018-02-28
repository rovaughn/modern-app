variable "domain" {
  type = "string"
}

variable "region" {
  type = "string"
}

variable "account_id" {
  type = "string"
}

provider "aws" {
  #region = "REGION"
}

terraform {
  backend "s3" {
    bucket = "BUCKET"
    key    = "terraform-state"

    #region = "us-east-1"
  }
}
