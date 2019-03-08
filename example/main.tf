variable "region" {
  description = "The AWS region where all the resources are created."
}

variable "account_id" {
  description = "The ID of the account in which all the resources are created."
}

provider "null" {
  version = "~> 1.0"
}

provider "template" {
  version = "~> 1.0"
}

provider "aws" {
  version = "~> 1.37"

  # Default region where everything is created.
  region = "${var.region}"

  # Define allowed account IDs to avoid touching other accounts by mistake.
  allowed_account_ids = ["${var.account_id}"]
}
