## Lambda function that provides the Gateway authorizer
resource "aws_lambda_function" "authorizer_lambda" {
  description      = "Lambda Function authorizing requests to the Pennsieve API v2."
  function_name    = "${var.environment_name}-${var.service_name}-authorizer_lambda-${data.terraform_remote_state.region.outputs.aws_region_shortname}"
  handler          = "authorizer_lambda"
  runtime          = "go1.x"
  role             = aws_iam_role.authorizer_lambda_role.arn
  timeout          = 300
  memory_size      = 128
  source_code_hash = data.archive_file.authorizer_lambda_archive.output_base64sha256
  filename         = "${path.module}/../lambda/bin/authorizer_lambda.zip"

  vpc_config {
    subnet_ids         = tolist(data.terraform_remote_state.vpc.outputs.private_subnet_ids)
    security_group_ids = [data.terraform_remote_state.platform_infrastructure.outputs.upload_v2_security_group_id]
  }

  environment {
    variables = {
      ENV = var.environment_name
      PENNSIEVE_DOMAIN = data.terraform_remote_state.account.outputs.domain_name,
      REGION = var.aws_region
      USER_POOL = data.terraform_remote_state.authentication_service.outputs.user_pool_2_id,
      USER_CLIENT = data.terraform_remote_state.authentication_service.outputs.user_pool_2_client_id,
      TOKEN_POOL = data.terraform_remote_state.authentication_service.outputs.token_pool_id,
      TOKEN_CLIENT = data.terraform_remote_state.authentication_service.outputs.token_pool_client_id
    }
  }
}

data "archive_file" "authorizer_lambda_archive" {
  type        = "zip"
  source_dir  = "${path.module}/../lambda/bin/authorizer"
  output_path = "${path.module}/../lambda/bin/authorizer_lambda.zip"
}
