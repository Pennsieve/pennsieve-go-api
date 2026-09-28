# T1 CloudWatch alarms (EPIC 868m2zvjt; standard sets from
# pennsieve-infra-dashboard/docs/alarm-coverage-plan.md). The authorizers
# gate every API request, so Errors/Throttles here are platform-wide
# incidents. No alarm_actions yet: alarms surface on the infra dashboard
# and console without paging.
module "service_alarms" {
  source = "git@github.com:Pennsieve/terraform-modules.git//service-alarms"

  environment_name = var.environment_name
  service_name     = var.service_name

  lambdas = {
    authorizer = {
      function_name   = aws_lambda_function.authorizer_lambda.function_name
      timeout_seconds = aws_lambda_function.authorizer_lambda.timeout
    }
    direct-authorizer = {
      function_name   = aws_lambda_function.direct_authorizer_lambda.function_name
      timeout_seconds = aws_lambda_function.direct_authorizer_lambda.timeout
    }
    websocket-authorizer = {
      function_name   = aws_lambda_function.websocket_authorizer_lambda.function_name
      timeout_seconds = aws_lambda_function.websocket_authorizer_lambda.timeout
    }
    events-authorizer = {
      function_name   = aws_lambda_function.events_authorizer_lambda.function_name
      timeout_seconds = aws_lambda_function.events_authorizer_lambda.timeout
    }
  }
}
