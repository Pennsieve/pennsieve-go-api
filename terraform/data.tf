data "aws_caller_identity" "current" {}

data "aws_region" "current_region" {}

# Import Account Data
data "terraform_remote_state" "account" {
  backend = "s3"

  config = {
    bucket = "${var.aws_account}-terraform-state"
    key    = "aws/terraform.tfstate"
    region = "us-east-1"
  }
}

# Import Upload-Service-v2
data "terraform_remote_state" "upload-service" {
  backend = "s3"

  config = {
    bucket = "${var.aws_account}-terraform-state"
    key    = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/upload-service-v2/terraform.tfstate"
    region = "us-east-1"
  }
}

# Import Authentication-Service
data "terraform_remote_state" "authentication_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/authentication-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}