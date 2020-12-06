package common

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

func UploadSpot(ctx context.Context, spot Spot, db dynamodbiface.DynamoDBAPI, tableName string) error {

	log.Println("UploadSpot")
	item, err := dynamodbattribute.MarshalMap(spot)
	if err != nil {
		log.Println("Error marshal map", err)
		return err
	}

	conditionExpression := "attribute_not_exists(SK)"
	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                item,
		ConditionExpression: aws.String(conditionExpression),
	})
	if err != nil {
		log.Println("Error putting item", err)
		return err
	}

	return nil

}

func GetSpotsWithGeohash(ctx context.Context, geohash string, db dynamodbiface.DynamoDBAPI, tableName string) ([]Spot, error) {

	log.Println("GetSpotsWithGeohash")

	sk := fmt.Sprintf("%s%s", SpotPrefix, geohash)
	keyConditionExpression := "#gsi2 = :gsi2 AND begins_with(#sk, :sk)"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi2": {S: aws.String("spots")},
		":sk":   {S: aws.String(sk)},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi2": aws.String("GSI2"),
		"#sk":   aws.String("SK"),
	}

	output, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("GSI2"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		return nil, err
	}

	spots := make([]Spot, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var spot Spot
		err := dynamodbattribute.UnmarshalMap(item, &spot)
		if err != nil {
			return nil, err
		}
		spots[index] = spot
	}
	return spots, nil
}

func AddSpotDistance(ctx context.Context, spotDistance SpotDistance, db dynamodbiface.DynamoDBAPI, tableName string) error {

	//log.Println("AddSpotDistance")
	item, err := dynamodbattribute.MarshalMap(spotDistance)
	if err != nil {
		log.Println("Error marshal map", err)
		return err
	}

	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		log.Println("Error putting item", err)
		return err
	}

	return nil
}

func GetSpotDistances(ctx context.Context, origin Spot, destinations []Spot, db dynamodbiface.DynamoDBAPI, tableName string) ([]SpotDistance, error) {

	var spotDistance []SpotDistance

	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk": {S: aws.String(origin.PK)},
	}
	expressionAttributeNames := map[string]*string{
		"#pk": aws.String("PK"),
		"#sk": aws.String("SK"),
	}

	sks := make([]string, len(destinations))
	for i, d := range destinations {
		skValueName := fmt.Sprintf(":sk%d", i)
		sks[i] = skValueName
		expressionAttributeValues[skValueName] = &dynamodb.AttributeValue{S: aws.String(d.PK)}
	}

	keyConditionExpression := fmt.Sprintf("#pk = :pk AND #sk IN(%s)", strings.Join(sks, ","))
	output, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		return nil, err
	}

	spotDistance = make([]SpotDistance, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var sd SpotDistance
		err := dynamodbattribute.UnmarshalMap(item, &sd)
		if err != nil {
			return nil, err
		}
		spotDistance[index] = sd
	}

	return spotDistance, nil
}
