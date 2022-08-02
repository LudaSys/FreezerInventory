package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       "Test",
		},
		nil
}

func main() {
	runtime.Start(handleRequest)
}
