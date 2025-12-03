# Welyo Recruitment Task - Http Server with CI/CD

## Overview

Simple HTTP server application developed using Golang, containers and CI/CD. The application exposes two main endpoints: a health check endpoint for monitoring and load balancer integration, and a hello-world endpoint that demonstrates environment variable configuration.

## API Endpoints

### GET /health-check

Returns a JSON response with a status of "ok".

Response:
```json
{
    "status": "ok"
}
```

### GET /hello-world

Returns a plain text response with the value of the SERVER_HELLO environment variable.

Response:
```text
HELLO WORLD
```

## Docker Health Checks

The Dockerfile includes a health check that runs the `/health-check` endpoint every 30 seconds. If the endpoint returns a status 200, the container is considered healthy. If the endpoint returns a different status, the container is considered unhealthy. The health check is implemented both in dockerfile and docker-compose.yml so that there always is a basic health check in place regardless of deployment method, but can be easily overwritten without rebuilding the image.

## Local Development

### Without Docker

To run the application locally, use the following command:

```bash
make run
```

### With Docker

To run the application locally using Docker, use the following command:

```bash
make docker-run
```

This command will automatically build the Docker image if it doesn't exist and then run the container. If you want to only build the image without running it:

```bash
make docker-build
```

You can also use Docker Compose to build and run the container:

```bash
docker compose up --build
```

## CI/CD

The CI/CD pipeline is defined in the `.github/workflows` directory. It includes two workflows:

1. `build-push.yml`: This workflow builds the Docker image and pushes it to GitHub Container Registry. It is triggered on every push event to the repository. Before the build & push, it runs code quality checks, including formatting, static analysis, and unit tests so that any bugged images won't be pushed to the registry.
2. `code-quality.yml`: This workflow runs code quality checks, including formatting, static analysis, and unit tests. It is triggered on every pull request event to the repository. It fails if any of the checks fail.
