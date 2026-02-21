output "execution_arn" {
  value = aws_apigatewayv2_api.upload-service-gateway.execution_arn
}

output "authorizer_invocation_role" {
  value = aws_iam_role.invocation_role.arn
}

output "authorizer_lambda_invoke_uri" {
  value = aws_lambda_function.authorizer_lambda.invoke_arn
}

output "direct_authorizer_lambda_arn" {
  value       = aws_lambda_function.direct_authorizer_lambda.arn
  description = "ARN of the direct authorizer Lambda function"
}

output "direct_authorizer_lambda_invoke_arn" {
  value       = aws_lambda_function.direct_authorizer_lambda.invoke_arn
  description = "Invoke ARN of the direct authorizer Lambda (for API Gateway integrations)"
}

output "direct_authorizer_lambda_name" {
  value       = aws_lambda_function.direct_authorizer_lambda.function_name
  description = "Name of the direct authorizer Lambda function"
}