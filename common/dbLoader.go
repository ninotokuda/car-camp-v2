package common

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/mmcloughlin/geohash"
)

func UploadSpot(ctx context.Context, spot Spot, db dynamodbiface.DynamoDBAPI, tableName string) error {

	LogInfo(ctx, "Invoke", "UploadSpot", nil)
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
		LogError(ctx, "Failed to put item", "UploadSpot", err, nil)
		return err
	}

	return nil

}

func GetSpotsWithGeohash(ctx context.Context, geohash string, db dynamodbiface.DynamoDBAPI, tableName string) ([]Spot, error) {

	LogInfo(ctx, "Invoke", "GetSpotsWithGeohash", nil)
	sk := fmt.Sprintf("%s%s", SpotPrefix, geohash)
	keyConditionExpression := "#gsi2 = :gsi2 AND begins_with(#sk, :sk)"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi2": {S: aws.String(SpotQueryName)},
		":sk":   {S: aws.String(sk)},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi2": aws.String("GSI2"),
		"#sk":   aws.String("SK"),
	}
	LogInfo(ctx, "Query", "GetSpotsWithGeohash", map[string]interface{}{
		"keyConditionExpression":    keyConditionExpression,
		"expressionAttributeValues": expressionAttributeValues,
		"expressionAttributeNames":  expressionAttributeNames,
	})

	output, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("GSI2"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		LogError(ctx, "Failed to query spots", "GetSpotsWithGeohash", err, nil)
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

func GetAllSpots(ctx context.Context, lastEvaluatedKey string, db dynamodbiface.DynamoDBAPI, tableName string) ([]Spot, string, error) {

	LogInfo(ctx, "Invoke", "GetAllSpots", nil)
	keyConditionExpression := "#gsi2 = :gsi2"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi2": {S: aws.String("spots")},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi2": aws.String("GSI2"),
	}
	LogInfo(ctx, "Query", "GetAllSpots", map[string]interface{}{
		"keyConditionExpression":    keyConditionExpression,
		"expressionAttributeValues": expressionAttributeValues,
		"expressionAttributeNames":  expressionAttributeNames,
	})

	queryInput := dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("GSI2"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	}
	if lastEvaluatedKey != "" {
		queryInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(lastEvaluatedKey)},
		}
	}

	output, err := db.Query(&queryInput)
	if err != nil {
		LogError(ctx, "Failed to query spots", "GetAllSpots", err, nil)
		return nil, "", err
	}

	spots := make([]Spot, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var spot Spot
		err := dynamodbattribute.UnmarshalMap(item, &spot)
		if err != nil {
			return nil, "", err
		}
		spots[index] = spot
	}
	var newLastEvaluatedKey string
	if output.LastEvaluatedKey != nil {
		if val, ok := output.LastEvaluatedKey["PK"]; ok {
			newLastEvaluatedKey = *val.S
		}
	}

	return spots, newLastEvaluatedKey, nil
}

func AddSpotDistance(ctx context.Context, spotDistance SpotDistance, db dynamodbiface.DynamoDBAPI, tableName string) error {

	LogInfo(ctx, "Invoke", "AddSpotDistance", nil)
	item, err := dynamodbattribute.MarshalMap(spotDistance)
	if err != nil {
		LogError(ctx, "Failed to marshal spotDistance", "AddSpotDistance", err, nil)
		return err
	}

	_, err = db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		LogError(ctx, "Failed to put item", "AddSpotDistance", err, nil)
		return err
	}

	return nil
}

func GetSpotDistances(ctx context.Context, origin Spot, destinations []Spot, db dynamodbiface.DynamoDBAPI, tableName string) ([]SpotDistance, error) {

	LogInfo(ctx, "Invoke", "GetSpotDistances", nil)
	var spotDistance []SpotDistance
	keyConditionExpression := fmt.Sprintf("#pk = :pk AND #sk = :sk")
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk": {S: aws.String(origin.PK)},
		":sk": {S: aws.String(SpotDistancesQueryName)},
	}
	expressionAttributeNames := map[string]*string{
		"#pk":   aws.String("PK"),
		"#sk":   aws.String("SK"),
		"#gsi1": aws.String("GSI1"),
	}

	gsi1s := make([]string, len(destinations))
	for i, d := range destinations {
		gsi1ValueValueName := fmt.Sprintf(":gsi1_%d", i)
		gsi1s[i] = gsi1ValueValueName
		expressionAttributeValues[gsi1ValueValueName] = &dynamodb.AttributeValue{S: aws.String(d.PK)}
	}
	filterExpression := fmt.Sprintf("#gsi1 IN(%s)", strings.Join(gsi1s, ","))
	LogInfo(ctx, "Query", "GetSpotDistances", map[string]interface{}{
		"keyConditionExpression":    keyConditionExpression,
		"expressionAttributeValues": expressionAttributeValues,
		"expressionAttributeNames":  expressionAttributeNames,
		"filterExpression":          filterExpression,
	})
	output, err := db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
		FilterExpression:          aws.String(filterExpression),
	})
	if err != nil {
		LogError(ctx, "Failed to get spotDistances", "GetSpotDistances", err, nil)
		return nil, err
	}

	spotDistance = make([]SpotDistance, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var sd SpotDistance
		err := dynamodbattribute.UnmarshalMap(item, &sd)
		if err != nil {
			LogError(ctx, "Failed to unmarshal spotDistance", "GetSpotDistances", err, nil)
			return nil, err
		}
		spotDistance[index] = sd
	}

	return spotDistance, nil
}

func CreateSpotDistances(ctx context.Context, spot Spot, db dynamodbiface.DynamoDBAPI, tableName string, mbClient MapboxClient) error {

	LogInfo(ctx, "Invoke", "CreateSpotDistances", nil)
	// get nearby geohashes
	nearbySpots := []Spot{}
	ghash4 := spot.Geohash()[:4]
	ghash4s := geohash.Neighbors(ghash4)

	for _, g4 := range ghash4s {
		ns, err := GetSpotsWithGeohash(ctx, g4, db, tableName)
		if err != nil {
			LogError(ctx, "Failed to get nearby Spots", "CreateSpot", err, nil)
			continue
		}
		nearbySpots = append(nearbySpots, ns...)
	}

	spotsInRange := []Spot{}
	for j := range nearbySpots {
		ns := nearbySpots[j]
		if ns.PK == spot.PK {
			continue
		}
		if Distance(spot.Latitude, spot.Longitude, ns.Latitude, ns.Longitude) <= 10000 {
			spotsInRange = append(spotsInRange, ns)
		}
	}

	if len(spotsInRange) == 0 {
		return nil
	}

	// check if distance already loaded
	existingSpotDistances, err := GetSpotDistances(ctx, spot, spotsInRange, db, tableName)
	if err != nil {
		LogError(ctx, "Failed get existing spot distances", "CreateSpot", err, nil)
		return err
	}

	noDistancesSpots := []Spot{}
	for j := range spotsInRange {
		sr := spotsInRange[j]
		hasSpotDistance := false
		for _, esd := range existingSpotDistances {
			if sr.SpotId() == esd.DestinationSpotId() {
				hasSpotDistance = true
				break
			}
		}
		if !hasSpotDistance {
			noDistancesSpots = append(noDistancesSpots, sr)
		}
	}

	if len(noDistancesSpots) == 0 {
		return nil
	}

	spotGroupSize := 24
	var spotGroups [][]Spot
	if len(noDistancesSpots)%spotGroupSize == 0 {
		spotGroups = make([][]Spot, len(noDistancesSpots)/spotGroupSize)
	} else {
		spotGroups = make([][]Spot, len(noDistancesSpots)/spotGroupSize+1)
	}

	for j := 0; j < len(spotGroups); j++ {
		end := int(math.Min(float64(len(noDistancesSpots)), float64((j+1)*spotGroupSize)))
		spotGroups[j] = noDistancesSpots[j*spotGroupSize : end]
	}
	for j := range spotGroups {
		sg := spotGroups[j]
		resp, err := mbClient.LoadDistances(ctx, spot, sg)
		if err != nil {
			LogError(ctx, "Failed to Load distances", "CreateSpot", err, nil)
			continue
		}

		distanceSpots := append([]Spot{spot}, sg...)
		for originIndex, origin := range distanceSpots {
			for destinationIndex, destination := range distanceSpots {
				if originIndex != destinationIndex {
					distanceSeconds := resp.Durations[originIndex][destinationIndex]
					distanceMeters := resp.Distances[originIndex][destinationIndex]
					spotDistance := NewSpotDistance(origin, destination, distanceSeconds, distanceMeters)
					err := AddSpotDistance(ctx, spotDistance, db, tableName)
					if err != nil {
						LogError(ctx, "Failed to add SpotDistance", "CreateSpot", err, nil)
					}
				}
			}
		}
	}

	return nil
}
