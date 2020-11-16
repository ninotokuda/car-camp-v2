package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	uuid "github.com/satori/go.uuid"
)

type SpotArgs struct {
	Pk              string
	Sk              string
	TalentId        string
	CompanyId       string
	Date            string
	DurationSeconds int32
	PriceYen        int32
	TalentName      *string
	Name            *string
	Description     *string
	StartDate       *string
	EndDate         *string
	UserId          string
	UserName        *string
	Status          *string
}

func (r *Resolver) OpenTalentSpots(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	log.Println("OpenTalentSpots")
	pk := fmt.Sprintf("spot#%s", args.TalentId)
	var keyConditionExpression string
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk":     {S: &pk},
		":status": {S: aws.String("open")},
	}
	// has range
	if args.StartDate != nil && args.EndDate != nil {
		keyConditionExpression = "GSI1 = :pk AND SK BETWEEN :startDate AND :endDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
	} else if args.StartDate != nil { // has start Date
		keyConditionExpression = "GSI1 = :pk AND SK >= :startDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
	} else if args.EndDate != nil { // has end Date
		keyConditionExpression = "PGSI1 = :pk AND SK <= :endDate"
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
	} else {
		keyConditionExpression = "GSI1 = :pk"
	}

	spotResolvers := []*SpotResolver{}
	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		IndexName:                 aws.String("GSI1"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		FilterExpression:          aws.String("#status = :status"),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("Status"),
		},
	})
	if err != nil {
		return nil, err
	}

	for index := range output.Items {
		item := output.Items[index]
		spot := spotFromDynamodbOutput(item)
		spotResolver := &SpotResolver{s: &spot}
		spotResolvers = append(spotResolvers, spotResolver)
	}

	return spotResolvers, nil

}

// change to use sk instead
func (r *Resolver) CompanySpots(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	log.Println("CompanySpots")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}
	pk := fmt.Sprintf("spot#%s", args.CompanyId)
	sellerPk := fmt.Sprintf("spot#%s", requestUser.CompanyId())
	if sellerPk != pk {
		log.Print("Error: user tries to make for company which he does not have access to")
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}

	var keyConditionExpression string
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk": {S: &pk},
	}
	expressionAttributeNames := map[string]*string{
		"#pk": aws.String("PK"),
	}
	// has range
	if args.StartDate != nil && args.EndDate != nil {
		keyConditionExpression = "#pk = :pk AND #date BETWEEN :startDate AND :endDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
		expressionAttributeNames["#date"] = aws.String("Date")
	} else if args.StartDate != nil { // has start Date
		keyConditionExpression = "#pk = :pk AND #date >= :startDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
		expressionAttributeNames["#date"] = aws.String("Date")
	} else if args.EndDate != nil { // has end Date
		keyConditionExpression = "#pk = :pk AND #date <= :endDate"
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
		expressionAttributeNames["#date"] = aws.String("Date")
	} else {
		keyConditionExpression = "#pk = :pk"
	}

	var filterExpression *string
	if args.Status != nil {
		filterExpression = aws.String("#status = :status")
		expressionAttributeValues[":status"] = &dynamodb.AttributeValue{S: args.Status}
		expressionAttributeNames["#status"] = aws.String("Status")
	}

	spotResolvers := []*SpotResolver{}
	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		IndexName:                 aws.String("PK-date-index"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
		FilterExpression:          filterExpression,
	})
	if err != nil {
		return nil, err
	}

	for index := range output.Items {
		item := output.Items[index]
		spot := spotFromDynamodbOutput(item)
		spotResolver := &SpotResolver{s: &spot}
		spotResolvers = append(spotResolvers, spotResolver)
	}

	return spotResolvers, nil

}

// change to use sk instead
func (r *Resolver) UserSpots(ctx context.Context, args SpotArgs) ([]*SpotResolver, error) {

	log.Println("UserSpots")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	if requestUser.UserId() != args.UserId {
		log.Print("Error: requestUser and request userid do not match ", requestUser.UserId(), "  -- ", args.UserId)
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}
	gsi2 := fmt.Sprintf("spot#user_%s", args.UserId)

	var keyConditionExpression string
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":gsi2": {S: &gsi2},
	}
	expressionAttributeNames := map[string]*string{
		"#gsi2": aws.String("GSI2"),
	}
	// has range
	if args.StartDate != nil && args.EndDate != nil {
		keyConditionExpression = "#gsi2 = :gsi2 AND #sk BETWEEN :startDate AND :endDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
		expressionAttributeNames["#sk"] = aws.String("SK")
	} else if args.StartDate != nil { // has start Date
		keyConditionExpression = "#gsi2 = :gsi2 AND #sk >= :startDate"
		expressionAttributeValues[":startDate"] = &dynamodb.AttributeValue{S: args.StartDate}
		expressionAttributeNames["#sk"] = aws.String("SK")
	} else if args.EndDate != nil { // has end Date
		keyConditionExpression = "#gsi2 = :gsi2 AND #sk <= :endDate"
		expressionAttributeValues[":endDate"] = &dynamodb.AttributeValue{S: args.EndDate}
		expressionAttributeNames["#sk"] = aws.String("SK")
	} else {
		keyConditionExpression = "#gsi2 = :gsi2"
	}

	var filterExpression *string
	if args.Status != nil {
		filterExpression = aws.String("#status = :status")
		expressionAttributeNames["#status"] = aws.String("Status")
		expressionAttributeValues[":status"] = &dynamodb.AttributeValue{S: args.Status}
	}

	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(r.TableName),
		IndexName:                 aws.String("GSI2"),
		KeyConditionExpression:    aws.String(keyConditionExpression),
		ExpressionAttributeValues: expressionAttributeValues,
		ExpressionAttributeNames:  expressionAttributeNames,
		FilterExpression:          filterExpression,
	})
	if err != nil {
		return nil, err
	}

	spotResolvers := []*SpotResolver{}
	for index := range output.Items {
		item := output.Items[index]
		spot := spotFromDynamodbOutput(item)
		spotResolver := &SpotResolver{s: &spot}
		spotResolvers = append(spotResolvers, spotResolver)
	}

	return spotResolvers, nil

}

func (r *Resolver) Spot(ctx context.Context, args TalentArgs) (*SpotResolver, error) {

	log.Println("Spot")
	output, err := r.Db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(args.Pk),
			},
			"SK": {
				S: aws.String(args.Sk),
			},
		},
	})

	if err != nil {
		return nil, err
	}

	item := output.Item
	spot := spotFromDynamodbOutput(item)
	spotResolver := &SpotResolver{s: &spot}

	return spotResolver, nil

}

func (r *Resolver) CreateSpot(ctx context.Context, args SpotArgs) (*SpotResolver, error) {

	log.Println("CreateSpot")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	if !requestUser.IsSellerUser() {
		log.Print("Error: requestUser is not seller")
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}
	spotId := fmt.Sprintf("spot_%s", uuid.NewV4().String())
	pk := fmt.Sprintf("spot#%s", args.CompanyId)
	sk := fmt.Sprintf("%s#%s", args.Date, spotId)
	gsi1 := fmt.Sprintf("spot#%s", args.TalentId)

	spot := Spot{
		PK:              pk,
		SK:              sk,
		SpotId:          spotId,
		CompanyId:       args.CompanyId,
		Status:          SpotStatusOpen,
		Date:            args.Date,
		DurationSeconds: args.DurationSeconds,
		PriceYen:        args.PriceYen,
		TalentName:      args.TalentName,
		Name:            args.Name,
		Description:     args.Description,
		GSI1:            aws.String(gsi1),
	}

	item, err := dynamodbattribute.MarshalMap(spot)
	if err != nil {
		log.Println("Error marshal map", err)
		return nil, err
	}
	output, err := r.Db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      item,
	})
	if err != nil {
		log.Println("Error putting item", err)
		return nil, err
	}
	log.Println("Output", output)

	spotResolver := &SpotResolver{s: &spot}

	return spotResolver, nil

}

func (r *Resolver) UpdateSpot(ctx context.Context, args SpotArgs) (*SpotResolver, error) {

	log.Println("UpdateSpot")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	companyId := requestUser.CompanyId()
	if fmt.Sprintf("spot#%s", companyId) != args.Pk {
		log.Printf("Error: request user company_id: %s does not match requested company_id: %s", companyId, args.Pk)
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}

	updateExpression := "SET #durationSeconds = :durationSeconds, #priceYen = :priceYen"
	expressionAttributeNames := map[string]*string{
		"#durationSeconds": aws.String("DurationSeconds"),
		"#priceYen":        aws.String("PriceYen"),
	}
	durationSecondsString := strconv.Itoa(int(args.DurationSeconds))
	priceYenString := strconv.Itoa(int(args.PriceYen))
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":durationSeconds": {N: &durationSecondsString},
		":priceYen":        {N: &priceYenString},
	}

	if args.Name != nil {
		updateExpression += ", #name = :name"
		expressionAttributeNames["#name"] = aws.String("Name")
		expressionAttributeValues[":name"] = &dynamodb.AttributeValue{S: args.Name}
	}

	if args.Description != nil {
		updateExpression += ", #description = :description"
		expressionAttributeNames["#description"] = aws.String("Description")
		expressionAttributeValues[":description"] = &dynamodb.AttributeValue{S: args.Name}
	}

	output, err := r.Db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: aws.String(args.Pk)},
			"SK": {S: aws.String(args.Sk)},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String("ALL_NEW"),
	})
	if err != nil {
		return nil, err
	}

	spot := spotFromDynamodbOutput(output.Attributes)
	spotResolver := &SpotResolver{s: &spot}
	return spotResolver, nil

}

func (r *Resolver) ReserveSpot(ctx context.Context, args SpotArgs) (*SpotResolver, error) {

	log.Println("ReserveSpot")

	// reserve spot
	pk := args.Pk
	sk := args.Sk
	gsi2 := fmt.Sprintf("spot#%s", args.UserId)
	userName := args.UserName
	status := SpotStatusReserved
	output, err := r.Db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {S: &pk},
			"SK": {S: &sk},
		},
		UpdateExpression:    aws.String("SET #gsi2 = :gsi2, #userName = :userName, #status = :status"),
		ConditionExpression: aws.String("attribute_not_exists(#gsi2)"), // only update is user does not exist
		ExpressionAttributeNames: map[string]*string{
			"#gsi2":     aws.String("GSI2"),
			"#userName": aws.String("UserName"),
			"#status":   aws.String("Status"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi2":     {S: &gsi2},
			":userName": {S: userName},
			":status":   {S: &status},
		},
		ReturnValues: aws.String("ALL_NEW"),
	})
	if err != nil {
		log.Println("Error updating table", err)
		return nil, err
	}

	spot := spotFromDynamodbOutput(output.Attributes)
	spotResolver := &SpotResolver{s: &spot}

	return spotResolver, nil

}

type Spot struct {
	PK              string
	SK              string
	SpotId          string
	CompanyId       string
	Status          string
	Date            string
	DurationSeconds int32
	PriceYen        int32
	TalentName      *string `json:"TalentName,omitempty"`
	UserName        *string `json:"UserName,omitempty"`
	Name            *string `json:"Name,omitempty"`
	Description     *string `json:"Description,omitempty"`
	GSI1            *string `json:"GSI1,omitempty"`
	GSI2            *string `json:"GSI2,omitempty"`
}

type SpotResolver struct {
	s *Spot
}

func (u *SpotResolver) PK(ctx context.Context) string {
	return u.s.PK
}

func (u *SpotResolver) SK(ctx context.Context) string {
	return u.s.SK
}

func (u *SpotResolver) SpotId(ctx context.Context) string {
	return u.s.SpotId
}

func (u *SpotResolver) CompanyId(ctx context.Context) string {
	return u.s.CompanyId
}

func (u *SpotResolver) Status(ctx context.Context) string {
	return u.s.Status
}

func (u *SpotResolver) Date(ctx context.Context) string {
	return u.s.Date
}

func (u *SpotResolver) DurationSeconds(ctx context.Context) int32 {
	return u.s.DurationSeconds
}

func (u *SpotResolver) PriceYen(ctx context.Context) int32 {
	return u.s.PriceYen
}

func (u *SpotResolver) TalentName(ctx context.Context) *string {
	return u.s.TalentName
}

func (u *SpotResolver) UserName(ctx context.Context) *string {
	return u.s.UserName
}

func (u *SpotResolver) Name(ctx context.Context) *string {
	return u.s.Name
}

func (u *SpotResolver) Description(ctx context.Context) *string {
	return u.s.Description
}

func (u *SpotResolver) GSI1(ctx context.Context) *string {
	return u.s.GSI1
}

func (u *SpotResolver) GSI2(ctx context.Context) *string {
	return u.s.GSI2
}

func spotFromDynamodbOutput(item map[string]*dynamodb.AttributeValue) Spot {
	log.Println("Spot from dynamodb output", item)
	var durationString, priceYenString int
	if row, ok := item["DurationSeconds"]; ok {
		if v, err := strconv.Atoi(*row.N); err == nil {
			durationString = v
		} else {
			log.Println("Error converting DurationSeconds into int", err)
		}
	}

	if row, ok := item["PriceYen"]; ok {
		if v, err := strconv.Atoi(*row.N); err == nil {
			priceYenString = v
		} else {
			log.Println("Error converting PriceYen into int", err)
		}
	}

	var name, description, gsi1, gsi2, userName, talentName *string
	if v, ok := item["Name"]; ok {
		name = v.S
	}
	if v, ok := item["Description"]; ok {
		description = v.S
	}
	if v, ok := item["GSI1"]; ok {
		gsi1 = v.S
	}
	if v, ok := item["GSI2"]; ok {
		gsi2 = v.S
	}
	if v, ok := item["UserName"]; ok {
		userName = v.S
	}
	if v, ok := item["TalentName"]; ok {
		talentName = v.S
	}
	spot := Spot{
		PK:              *item["PK"].S,
		SK:              *item["SK"].S,
		SpotId:          *item["SpotId"].S,
		Status:          *item["Status"].S,
		Date:            *item["Date"].S,
		DurationSeconds: int32(durationString),
		PriceYen:        int32(priceYenString),
		Name:            name,
		Description:     description,
		GSI1:            gsi1,
		GSI2:            gsi2,
		UserName:        userName,
		TalentName:      talentName,
	}
	return spot
}
