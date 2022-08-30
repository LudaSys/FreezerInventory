package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	_ "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"log"
	"strconv"
)

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

type FoodItem struct {
	itemId string `json:"itemId"`
	Name   string `json:"name"`
	Time   string `json:"time"`
}

func (basics TableBasics) AddFoodItem(foodItem FoodItem) (bool, error) {
	_, err := basics.DynamoDbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("FoodItems"),
		Item: map[string]types.AttributeValue{
			"itemId": &types.AttributeValueMemberS{Value: foodItem.itemId},
			"name":   &types.AttributeValueMemberS{Value: foodItem.Name},
			"time":   &types.AttributeValueMemberS{Value: "12345"},
		},
	})
	if err != nil {
		log.Printf("Couldn't add item to table. Here's why: %v\n", err)
	}
	return false, err
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(o *config.LoadOptions) error {
		o.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	result, err := TableBasics.AddFoodItem(TableBasics{
		DynamoDbClient: dynamodb.NewFromConfig(cfg),
		TableName:      "FoodItems",
	}, FoodItem{
		itemId: uuid.New().String(),
		Name:   "Banana",
	})

	return events.APIGatewayProxyResponse{
			StatusCode: 201,
			Body:       strconv.FormatBool(result),
		},
		err
}

func main() {
	runtime.Start(handleRequest)
}
