package main

import (
	"context"
	"encoding/json"
	"functions/shared/models"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}
	var requestModel models.FoodItem
	json.Unmarshal([]byte(request.Body), &requestModel)

	svc := dynamodb.NewFromConfig(cfg)
	_, err = svc.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("FoodItems"),
		Item: map[string]types.AttributeValue{
			"itemId":           &types.AttributeValueMemberS{Value: uuid.New().String()},
			"name":             &types.AttributeValueMemberS{Value: requestModel.Name},
			"storage_location": &types.AttributeValueMemberS{Value: requestModel.StorageLocation},
		},
	})

	if err != nil {
		panic(err)
	}

	return events.APIGatewayProxyResponse{
			StatusCode: 201,
		},
		err
}

func main() {
	runtime.Start(handleRequest)
}
