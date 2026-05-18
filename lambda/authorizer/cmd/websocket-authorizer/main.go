// Package main is the Lambda entry point for the WebSocket API Gateway
// REQUEST authorizer.
//
// See lambda/authorizer/handler/websocket_handler.go for the full rationale on
// why this is a separate Lambda from the HTTP `Handler`. tl;dr: API Gateway
// WebSocket REQUEST authorizers only support payload format 1.0, the existing
// `Handler` is hard-coded to payload format 2.0, and the two event shapes
// differ enough that splitting is cleaner than dispatching by shape.
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pennsieve/pennsieve-go-api/authorizer/handler"
)

func main() {
	lambda.Start(handler.WebSocketHandler)
}
