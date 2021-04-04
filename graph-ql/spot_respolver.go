package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ninotokuda/carcamp_v2/common"
	uuid "github.com/satori/go.uuid"
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

	common.LogInfo(ctx, "Invoke", "Spot", map[string]interface{}{"args": args})
	pk := fmt.Sprintf("%s%s", common.SpotPrefix, args.SpotId)
	keyConditionExpression := "#pk = :pk AND begins_with(#sk, :sk)"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk": {S: aws.String(pk)},
		":sk": {S: aws.String(SpotPrefix)},
	}
	expressionAttributeNames := map[string]*string{
		"#pk": aws.String("PK"),
		"#sk": aws.String("SK"),
	}

	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
		Limit:                     aws.Int64(1),
	})
	if err != nil {
		common.LogError(ctx, "Failed to query spot", "Spot", err, nil)
		return nil, err
	}

	if len(output.Items) != 1 {
		common.LogError(ctx, "Did not find spot", "Spot", nil, nil)
		return nil, errors.New("Did not find spot")
	}

	var spot common.Spot
	err = dynamodbattribute.UnmarshalMap(output.Items[0], &spot)
	if err != nil {
		common.LogError(ctx, "Failed to unmarshal spot", "Spot", err, nil)
		return nil, err
	}
	spotResolver := &SpotResolver{spot: &spot, baseResolver: r}
	return spotResolver, nil

}

// change to use sk instead
func (r *Resolver) SpotsByGeohash(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	logInfo(ctx, "Invoke", "SpotsByGeohash", map[string]interface{}{"args": args})
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

	logInfo(ctx, "Invoke", "SpotsByCreator", map[string]interface{}{"args": args})
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

type CreateSpotArgs struct {
	CreatorUserId string
	Goehash       string
	SpotType      string
	Latitude      float64
	Longitude     float64
	Name          *string
	Address       *string
	Code          *string
	Prefecture    *string
	City          *string
	HomePageUrls  *[]string
	Tags          *[]string
}

func (r *Resolver) CreateSpot(ctx context.Context, args CreateSpotArgs) (*SpotResolver, error) {

	logInfo(ctx, "Invoke", "CreateSpot", map[string]interface{}{"args": args})
	creationTime := time.Now().Format(time.RFC3339)
	spotId := uuid.NewV4().String()
	pk := fmt.Sprintf("%s%s", common.SpotPrefix, spotId)
	sk := fmt.Sprintf("%s%s", common.GeohashPrefix, args.Goehash)
	gsi1 := fmt.Sprintf("%s%s", common.UserPrefix, args.CreatorUserId)
	spot := common.Spot{
		PK:           pk,
		SK:           sk,
		GSI1:         aws.String(gsi1),
		GSI2:         aws.String(common.SpotQueryName),
		CreationTime: creationTime,
		SpotType:     args.SpotType,
		Latitude:     args.Latitude,
		Longitude:    args.Longitude,
		Name:         args.Name,
		Address:      args.Address,
		Code:         args.Code,
		Prefecture:   args.Prefecture,
		City:         args.City,
		HomePageUrls: args.HomePageUrls,
		Tags:         args.Tags,
	}

	item, err := dynamodbattribute.MarshalMap(spot)
	if err != nil {
		logError(ctx, "failed to marshal spot to map", "CreateSpot", err, nil)
		return nil, err
	}
	_, err = r.Db.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(r.TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(SK)"),
	})
	if err != nil {
		logError(ctx, "failed to put item", "CreateSpot", err, nil)
		return nil, err
	}

	// add to mapbox
	err = r.MapboxClient.AddFeature(ctx, spot)
	if err != nil {
		logError(ctx, "Failed to add feature to mapbox", "CreateSpot", err, nil)
		return nil, err
	}

	// create spot distances
	err = common.CreateSpotDistances(ctx, spot, r.Db, r.TableName, r.MapboxClient)
	if err != nil {
		logError(ctx, "Failed to create spot distances", "CreateSpot", err, nil)
		return nil, err
	}

	spotResolver := SpotResolver{spot: &spot, baseResolver: r}
	return &spotResolver, nil
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

func (z SpotResolver) HomePageUrls(ctx context.Context) *[]string {
	return z.spot.HomePageUrls
}

func (z SpotResolver) Tags(ctx context.Context) *[]string {
	return z.spot.Tags
}

func (z SpotResolver) DefaultImageUrl(ctx context.Context) *string {
	return z.spot.DefaultImageUrl
}
