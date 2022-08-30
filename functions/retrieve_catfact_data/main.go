package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"io/ioutil"
	"net/http"
)

func getData() (fact CatFact) {
	resp, err := http.Get("https://catfact.ninja/fact")

	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			defer resp.Body.Close()
			return
		}

		unMarshalErr := json.Unmarshal(body, &fact)

		if unMarshalErr != nil {
			return CatFact{}
		}
		return
	}

	return
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       getData().FactText,
		},
		nil
}

func main() {
	runtime.Start(handleRequest)
}
