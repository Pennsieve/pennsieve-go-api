##############################
# AUTHORIZER-INVOCATION-ROLE #
##############################

resource "aws_iam_role" "invocation_role" {
  name = "api_gateway_auth_invocation"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "default"
  role = aws_iam_role.invocation_role.id

  // Grants the API Gateway invocation role permission to call BOTH the HTTP
  // authorizer (used by REST + HTTP V2 APIs) and the WebSocket authorizer
  // (used by API Gateway V2 WebSocket APIs). The two Lambdas are separate
  // because WebSocket REQUEST authorizers only support payload format 1.0
  // while HTTP V2 authorizers use format 2.0 — see
  // lambda/authorizer/handler/websocket_handler.go for the long-form note.
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": [
        "${aws_lambda_function.authorizer_lambda.arn}",
        "${aws_lambda_function.websocket_authorizer_lambda.arn}"
      ]
    }
  ]
}
EOF
}

##############################
# UPLOAD-SERVICE-LAMBDA   #
##############################
// 1. Lambda can assume the upload_trigger_lambda role
// 2. This role has a policy attachment
// 3. This policy has a policy document attached
// 4. This document outlines the permissions for the role

resource "aws_iam_role" "authorizer_lambda_role" {
  name = "${var.environment_name}-${var.service_name}-authorizer-lambda-role-${data.terraform_remote_state.region.outputs.aws_region_shortname}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "authorizer_lambda_iam_policy_attachment" {
  role       = aws_iam_role.authorizer_lambda_role.name
  policy_arn = aws_iam_policy.authorizer_lambda_iam_policy.arn
}

resource "aws_iam_policy" "authorizer_lambda_iam_policy" {
  name   = "${var.environment_name}-${var.service_name}-authorizer-lambda-iam-policy-${data.terraform_remote_state.region.outputs.aws_region_shortname}"
  path   = "/"
  policy = data.aws_iam_policy_document.authorizer_lambda_iam_policy_document.json
}

data "aws_iam_policy_document" "authorizer_lambda_iam_policy_document" {

  statement {
    sid = "LambdaAccessToDynamoDB"
    effect = "Allow"

    actions = [
      "dynamodb:DescribeTable",
      "dynamodb:BatchGetItem",
      "dynamodb:GetItem",
      "dynamodb:Query",
      "dynamodb:Scan",
    ]

    resources = [
      data.terraform_remote_state.upload_service_v2.outputs.manifest_table_arn,
      "${data.terraform_remote_state.upload_service_v2.outputs.manifest_table_arn}/*",
      data.terraform_remote_state.upload_service_v2.outputs.manifest_table_arn,
      "${data.terraform_remote_state.upload_service_v2.outputs.manifest_table_arn}/*"
    ]

  }

  statement {
    sid    = "UploadLambdaPermissions"
    effect = "Allow"
    actions = [
      "rds-db:connect",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutDestination",
      "logs:PutLogEvents",
      "logs:DescribeLogStreams",
      "ec2:CreateNetworkInterface",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DeleteNetworkInterface",
      "ec2:AssignPrivateIpAddresses",
      "ec2:UnassignPrivateIpAddresses"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "InvokeCallbackValidatorLambdas"
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction"
    ]
    resources = [
      data.terraform_remote_state.workflow_service.outputs.callback_validator_lambda_arn,
    ]
  }

  // Authorizer Lambdas (specifically the websocket-authorizer) need to
  // invoke account-service's check-access Lambda when the WebSocket
  // handshake URL carries `?computeNodeId=...`. This is the first runtime
  // call from pennsieve-go-api outward into account-service — see the
  // package doc at the top of lambda/authorizer/handler/websocket_handler.go
  // and the matching env-var wiring in lambda.tf for the architectural
  // rationale.
  //
  // Resource is scoped to the specific check-access Lambda ARN; the
  // authorizer cannot invoke arbitrary functions in account-service.
  statement {
    sid    = "AuthorizerInvokeCheckAccess"
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction"
    ]
    resources = [
      data.terraform_remote_state.account_service.outputs.check_access_lambda_arn,
    ]
  }

}
