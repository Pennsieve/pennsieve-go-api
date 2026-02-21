package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pennsieve/pennsieve-go-api/authorizer/handler"
)

func main() {
	lambda.Start(handler.DirectHandler)
}
