output "execution_arn" {
  value = aws_apigatewayv2_api.upload-service-gateway.execution_arn
}

output "authorizer_invocation_role" {
  value = aws_iam_role.invocation_role.arn
}

output "authorizer_lambda_invoke_uri" {
  value = aws_lambda_function.authorizer_lambda.invoke_arn
}