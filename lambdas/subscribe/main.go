package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/aws/aws-lambda-go/lambda"
)

// Response from API
type Response struct {
	PhoneNumber string `json:"phoneNumber"`
}

// Request struct - incoming HTTP request
type Request struct {
	PhoneNumber string `json:"phoneNumber"`
}

// Handler function for lambda
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	//				START MONGODB SETUP					//
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		os.Getenv("MONGODB_URI"),
	))
	if err != nil {
		log.Fatal("Connection error:", err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Ping error:", err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	coronalertDB := client.Database("Coronalert")
	phoneNumbersCollection := coronalertDB.Collection("PhoneNumbers")
	//				END MONGODB SETUP					//

	//				START SUBSCRIBE						//
	bodyRequest := Request{
		PhoneNumber: "",
	}

	err = json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		log.Fatal("error in unmarshal")
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	_, err = phoneNumbersCollection.InsertOne(ctx, bson.D{
		{Key: "phoneNumber", Value: bodyRequest.PhoneNumber},
	})
	if err != nil {
		log.Fatal("error adding to collection")
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	bodyResponse := Response{
		PhoneNumber: bodyRequest.PhoneNumber,
	}

	response, err := json.Marshal(&bodyResponse)
	if err != nil {
		log.Fatal("error in marshal")
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}
	//				END SUBSCRIBE						//

	return events.APIGatewayProxyResponse{
		Body:       string(response),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
