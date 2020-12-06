package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ninotokuda/carcamp_v2/common"
)

type SpotArgs struct {
	SpotId       string
	CreatorId    string
	Geohash      string
	SpotTypes    *[]*string
	SpotType     string
	Latitude     float64
	Longitude    float64
	Name         string
	Address      string
	Code         string
	Prefecture   string
	City         string
	HomePageUrls []string
	Tags         []string
}

func (r *Resolver) Spot(ctx context.Context, args SpotArgs) (*SpotResolver, error) {

	log.Println("Spot")
	pk := fmt.Sprintf("spot#%s", args.SpotId)
	output, err := r.Db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(pk),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	var spot common.Spot
	err = dynamodbattribute.UnmarshalMap(output.Item, &spot)
	if err != nil {
		return nil, err
	}
	spotResolver := &SpotResolver{spot: &spot, baseResolver: r}
	return spotResolver, nil

}

// change to use sk instead
func (r *Resolver) SpotsByGeohash(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	log.Println("SpotsByGeohash")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	keyConditionExpression := "#gsi2 = :gsi2"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi2": {S: aws.String("spots")},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi2": aws.String("GSI2"),
	}

	geohash := fmt.Sprintf("%s%s", SpotPrefix, args.Geohash)
	expressionAttributeValues[":sk"] = &dynamodb.AttributeValue{S: aws.String(geohash)}
	expressionAttributeNames["#sk"] = aws.String("SK")

	keyConditionExpression = keyConditionExpression + " AND begins_with(#sk, :sk)"

	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		IndexName:                 aws.String("GSI2"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		return nil, err
	}

	spotResolvers := make([]*SpotResolver, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var spot common.Spot
		err := dynamodbattribute.UnmarshalMap(item, &spot)
		if err != nil {
			return nil, err
		}
		spotResolver := &SpotResolver{spot: &spot}
		spotResolvers[index] = spotResolver
	}

	return spotResolvers, nil

}

func (r *Resolver) SpotsByCreator(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	log.Println("Spots")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	gsi1 := fmt.Sprintf("%s%s", UserPrefix, args.CreatorId)
	keyConditionExpression := "#gsi1 = :gsi1"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi1": {S: aws.String(gsi1)},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi1": aws.String("GSI1"),
	}

	if args.SpotTypes != nil && len(*args.SpotTypes) > 0 {
		//TODO
	}

	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		IndexName:                 aws.String("GSI1"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
	})
	if err != nil {
		return nil, err
	}
	spotResolvers := make([]*SpotResolver, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var spot common.Spot
		err := dynamodbattribute.UnmarshalMap(item, &spot)
		if err != nil {
			return nil, err
		}
		spotResolver := &SpotResolver{spot: &spot}
		spotResolvers[index] = spotResolver
	}

	return spotResolvers, nil

}

type SpotResolver struct {
	spot         *common.Spot
	baseResolver *Resolver // consider using interface instead
}

func (z SpotResolver) SpotId(ctx context.Context) string {
	return strings.TrimPrefix(z.spot.PK, SpotPrefix)
}

func (z SpotResolver) Geohash(ctx context.Context) string {
	return strings.TrimPrefix(z.spot.PK, SpotPrefix)
}

func (z SpotResolver) SpotType(ctx context.Context) string {
	return z.spot.SpotType
}

func (z SpotResolver) Latitude(ctx context.Context) float64 {
	return z.spot.Latitude
}

func (z SpotResolver) Longitude(ctx context.Context) float64 {
	return z.spot.Longitude
}

func (z SpotResolver) CreationTime(ctx context.Context) string {
	return z.spot.CreationTime
}

func (z SpotResolver) Reviews(ctx context.Context) (*[]*ReviewResolver, error) {
	spotId := z.SpotId(ctx)
	reviewArgs := ReviewArgs{SpotId: aws.String(spotId)}
	reslovers, err := z.baseResolver.Reviews(ctx, reviewArgs)
	return &reslovers, err
}

func (z SpotResolver) SpotDistances(ctx context.Context) (*[]*SpotDistanceResolver, error) {

	return nil, nil
}

func (z SpotResolver) Images(ctx context.Context) (*[]*SpotImageResolver, error) {

	return nil, nil
}

func (z SpotResolver) CreatorId(ctx context.Context) *string {
	if z.spot.GSI1 != nil {
		return aws.String(strings.TrimPrefix(*z.spot.GSI1, UserPrefix))
	}
	return nil
}

func (z SpotResolver) Creator(ctx context.Context) (*UserResolver, error) {
	if userId := z.CreatorId(ctx); userId != nil {
		userArgs := UserArgs{UserId: *userId}
		return z.baseResolver.User(ctx, userArgs)
	}
	return nil, nil
}

func (z SpotResolver) Name(ctx context.Context) *string {
	return z.spot.Name
}

func (z SpotResolver) Description(ctx context.Context) *string {
	return z.spot.Description
}

func (z SpotResolver) Address(ctx context.Context) *string {
	return z.spot.Address
}

func (z SpotResolver) Code(ctx context.Context) *string {
	return z.spot.Code
}

func (z SpotResolver) Prefecture(ctx context.Context) *string {
	return z.spot.Prefecture
}

func (z SpotResolver) City(ctx context.Context) *string {
	return z.spot.City
}

func (z SpotResolver) HomePageUrls(ctx context.Context) *[]*string {
	l := z.spot.HomePageUrls
	return &l
}

func (z SpotResolver) Tags(ctx context.Context) *[]*string {
	l := z.spot.Tags
	return &l
}
