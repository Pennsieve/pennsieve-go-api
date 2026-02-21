// Create log group for authorizer lambda.
# resource "aws_cloudwatch_log_group" "authorizer_lambda_loggroup" {
#   name              = "/aws/lambda/${aws_lambda_function.authorizer_lambda.function_name}"
#   retention_in_days = 30
#   tags = local.common_tags
# }

// Send logs from authorizer lambda to datadog
// Currently the build fails if the block above is included as a log group already exists
// Once the state is fixed (i.e. manually created resources deleted),
// the code below should be updated to reference the log group created in the block above, and
// not reference the created lambda directly
# resource "aws_cloudwatch_log_subscription_filter" "cloudwatch_log_group_subscription" {
#   name            = "/aws/lambda/${aws_lambda_function.authorizer_lambda.function_name}-subscription"
#   log_group_name  = "/aws/lambda/${aws_lambda_function.authorizer_lambda.function_name}"
#   filter_pattern  = ""
#   destination_arn = data.terraform_remote_state.region.outputs.datadog_delivery_stream_arn
#   role_arn        = data.terraform_remote_state.region.outputs.cw_logs_to_datadog_logs_firehose_role_arn
# }

resource "aws_cloudwatch_log_group" "direct_authorizer_lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.direct_authorizer_lambda.function_name}"
  retention_in_days = 30
}
