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

# Import Region Data
data "terraform_remote_state" "region" {
  backend = "s3"

  config = {
    bucket = "${var.aws_account}-terraform-state"
    key    = "aws/${data.aws_region.current_region.name}/terraform.tfstate"
    region = "us-east-1"
  }
}

# Import VPC Data
data "terraform_remote_state" "vpc" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Platform Infrastructure Data
data "terraform_remote_state" "platform_infrastructure" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/platform-infrastructure/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Postgres
data "terraform_remote_state" "pennsieve_postgres" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/pennsieve-postgres/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Postgres
data "terraform_remote_state" "upload_service_v2" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/upload-service-v2/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Model-Service-Serverless
data "terraform_remote_state" "model_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/model-service-serverless/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Datasets Service
data "terraform_remote_state" "datasets_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/datasets-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Packages Service
data "terraform_remote_state" "packages_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/packages-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Imaging Service
data "terraform_remote_state" "imaging_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/imaging-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Service
data "terraform_remote_state" "import_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/import-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Integration Service
data "terraform_remote_state" "integration_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/integration-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Rehydration Service
data "terraform_remote_state" "rehydration_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/rehydration-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Readme Service
data "terraform_remote_state" "readme_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/readme-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Account Service
data "terraform_remote_state" "account_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/account-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Github  Service
data "terraform_remote_state" "github_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/github-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import Compute Node Service
data "terraform_remote_state" "compute_node_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/compute-node-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}

# Import App Deploy Service
data "terraform_remote_state" "app_deploy_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/app-deploy-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}
# DRS Service
data "terraform_remote_state" "drs_service" {
  backend = "s3"

  config = {
    bucket  = "${var.aws_account}-terraform-state"
    key     = "aws/${data.aws_region.current_region.name}/${var.vpc_name}/${var.environment_name}/drs-service/terraform.tfstate"
    region  = "us-east-1"
    profile = var.aws_account
  }
}
