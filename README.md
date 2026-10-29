# pennsieve-go-api
Golang API for the Pennsieve Platform

> ⚠️ **The API-Gateway authorizers live in `auth-service` now — do NOT change authorizer code in this repo.**
>
> The Lambda authorizers (HTTP / direct / WebSocket / events) were consolidated into
> [`Pennsieve/auth-service`](https://github.com/Pennsieve/auth-service) as part of the Auth Refactor
> (Phase 0). The authorizer code and Terraform that remain here under `lambda/authorizer/` and
> `terraform/` are pending decommission (workstream F2) and are kept only so the home gateway keeps
> working until it is fully cut over. **Any authorizer change — new scheme, claim, events/AppSync auth,
> bug fix — belongs in `auth-service`, not here.** Changes made here will be lost when these lambdas are
> retired, and will diverge from the authorizer that actually serves traffic.

## API Gateway
The Pennsieve-Go-API is a serverless api that is built around an AWS API Gateway.

The API Gateway routes traffic to the api to various Lambda functions that
are defined in separate services, which are manages in independent Github repositories.

## API Controllers and Models
The API provides interfaces with the Postgres DB.


## Testing

The tests are automatically run on Jenkins in a Docker container by `make test-ci` once you push to a feature branch. Successful tests are required to merge a feature branch into the main branch.

### Testing locally

Run `make test`. This will run the same tests as `make test-ci`, but they will be run directly by `go test` and not in a separate Docker container as happens with `make test-ci`.

If you want to run or debug individual tests in your IDE, first run `make local-services`. This will start the Docker containers required by some tests: a Postgres with the pennsieve-seed DB and an empty, local, in-memory DynamoBB.

## Deployment

__Build and Development Deployment__

Artifacts are built in Jenkins and published to S3. The dev build triggers a deployment of the Lambda function and creates a "Lambda version" that is used by the model-service.

__Deployment of an Artifact__

1. Deployments to *development* are automatically done by Jenkins once you merge a feature branch into main.

2. Deployments to *production* are done via Jenkins.

    1. Determine the artifact version you want to deploy (you can find the latest version number in the development deployment job).
    2. Run the production deployment task with the new `IMAGE_TAG`
    
Note: After terraforming the authorizer, you need to manually add the invoke role
to the authorizer as this is currently not automatically picked up from the OAS 
configuration for HTTP APIs. (⚠️ As noted at the top — the authorizers are being
decommissioned in favor of `auth-service`; avoid new authorizer work here.)
