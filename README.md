# pennsieve-go-api
Golang API for the Pennsieve Platform

## API Gateway
The Pennsieve-Go-API is a serverless api that is built around an AWS API Gateway.

The API Gateway routes traffic to the api to various Lambda functions that
are defined in separate services, which are manages in independent Github repositories.

## API Controllers and Models
The API provides interfaces with the Postgres DB.


## Lambda Authorizer

Note: After terraforming the authorizer, you need to manually add the invoke role
to the authorizer as this is currently not automatically picked up from the OAS 
configuration for HTTP APIs.

```env GOOS=linux GOARCH=amd64 go build -o ../bin/authorizer/authorizer_lambda```