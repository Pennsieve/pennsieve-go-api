// Package main is the Lambda entry point for the AppSync Event API
// authorizer. See lambda/authorizer/handler/events_handler.go.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pennsieve/pennsieve-go-api/authorizer/handler"
)

func main() {
	lambda.Start(handler.EventsHandler)
}
