// Create log group for authorizer lambda.
# resource "aws_cloudwatch_log_group" "authorizer_lambda_loggroup" {
#   name              = "/aws/lambda/${aws_lambda_function.authorizer_lambda.function_name}"
#   retention_in_days = 30
#   tags = local.common_tags
# }

resource "aws_cloudwatch_log_group" "direct_authorizer_lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.direct_authorizer_lambda.function_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "websocket_authorizer_lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.websocket_authorizer_lambda.function_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "events_authorizer_lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.events_authorizer_lambda.function_name}"
  retention_in_days = 30
}
