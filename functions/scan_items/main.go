package main

import (
	"context"
	"encoding/json"
	"fmt"
	"functions/shared/models"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	svc := dynamodb.NewFromConfig(cfg)

	out, err := svc.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String("FoodItems"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(out.Items)
	var results []models.FoodItem
	unmarhsalErr := attributevalue.UnmarshalListOfMaps(out.Items, &results)

	if unmarhsalErr != nil {
		return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Body:       "Error unmarshalling data",
			},
			nil
	}

	response, error := json.Marshal(results)

	return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       string(response),
		},
		error
}

func main() {
	runtime.Start(handleRequest)
}
