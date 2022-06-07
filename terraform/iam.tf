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

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.authorizer_lambda.arn}"
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



}
