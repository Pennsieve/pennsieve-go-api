variable "aws_account" {}

variable "aws_region" {}

variable "environment_name" {}

variable "service_name" {}

variable "vpc_name" {}

variable "domain_name" {}

variable "pennsieve_postgres_host" {}

variable "user_pool_endpoint" {
  default = "cognito-idp.us-east-1.amazonaws.com/us-east-1_FVLhJ7CQA"
}

variable "user_pool_2_client_id" {
  default = "703lm5d8odccu21pagcfjkeaea"
}

variable "token_pool_endpoint" {
  default = "cognito-idp.us-east-1.amazonaws.com/us-east-1_uCQXlh5nG"
}

variable "token_pool_client_id" {
  default = "p18fdvhilhj2tg5sahtcsh6m6"
}

variable "image_tag" {
}

variable "lambda_bucket" {
  default = "pennsieve-cc-lambda-functions-use1"
}

locals {
  domain_name = data.terraform_remote_state.account.outputs.domain_name
  hosted_zone = data.terraform_remote_state.account.outputs.public_hosted_zone_id

  common_tags = {
    aws_account      = var.aws_account
    aws_region       = data.aws_region.current_region.name
    environment_name = var.environment_name
  }
}
