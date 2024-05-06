resource "aws_apigatewayv2_api" "upload-service-gateway" {
  name          = "serverless_upload_service"
  protocol_type = "HTTP"
  description = "API Gateway for Upload-Service V2"
  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["*"]
    allow_headers = ["*"]
    expose_headers = ["*"]
    max_age = 300
  }
  body          = templatefile("${path.module}/upload_service.yml", {
    upload_service_lambda_arn = data.terraform_remote_state.upload-service.outputs.service_lambda_arn,
    model_service_lambda_arn = data.terraform_remote_state.model_service.outputs.service_lambda_arn,
    publishing_service_lambda_arn = data.terraform_remote_state.publishing_service.outputs.service_lambda_arn,
    datasets_service_lambda_arn = data.terraform_remote_state.datasets_service.outputs.service_lambda_arn,
    packages_service_lambda_arn = data.terraform_remote_state.packages_service.outputs.service_lambda_arn,
    integration_service_lambda_arn = data.terraform_remote_state.integration_service.outputs.lambda_service_arn,
    rehydration_service_lambda_arn = data.terraform_remote_state.rehydration_service.outputs.rehydration_service_arn,
    readme_service_lambda_arn = data.terraform_remote_state.readme_service.outputs.service_lambda_arn,
    account_service_lambda_arn = data.terraform_remote_state.account_service.outputs.service_lambda_arn,
    compute_node_service_lambda_arn = data.terraform_remote_state.compute_node_service.outputs.service_lambda_arn,
    user_pool_2_client_id = data.terraform_remote_state.authentication_service.outputs.user_pool_2_client_id,
    user_pool_endpoint = "https://${var.user_pool_endpoint}"
    token_pool_client_id = data.terraform_remote_state.authentication_service.outputs.token_pool_client_id,
    token_pool_endpoint = "https://${var.token_pool_endpoint}"
    authorize_lambda_invoke_uri = aws_lambda_function.authorizer_lambda.invoke_arn
    gateway_authorizer_role = aws_iam_role.invocation_role.arn
  })
}

resource "aws_apigatewayv2_stage" "upload-service-gateway-stage" {
  api_id = aws_apigatewayv2_api.upload-service-gateway.id

  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.upload-service-log-group.arn

    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
    }
    )
  }
}

resource "aws_apigatewayv2_integration" "int" {
  api_id           = aws_apigatewayv2_api.upload-service-gateway.id
  integration_type = "AWS_PROXY"
  connection_type = "INTERNET"
  integration_method = "POST"
  integration_uri = data.terraform_remote_state.upload-service.outputs.service_lambda_invoke_arn
}

resource "aws_cloudwatch_log_group" "upload-service-log-group" {
  name =  "${var.environment_name}/${var.service_name}/serverless_api_gateway"

  retention_in_days = 30
}

resource "aws_lambda_permission" "upload-service-lambda-permission" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = data.terraform_remote_state.upload-service.outputs.service_lambda_function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.upload-service-gateway.execution_arn}/*/*"
}