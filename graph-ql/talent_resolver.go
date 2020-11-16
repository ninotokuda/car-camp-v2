package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"

	_ "image/gif"
	"image/jpeg"
	"image/png"
)

func (r *Resolver) AllTalents(ctx context.Context) ([]*TalentResolver, error) {

	keyConditionExpression := "GSI1 = :gsi1"
	talentResolvers := []*TalentResolver{}
	gsi1 := "talent"
	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String(keyConditionExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1": {S: &gsi1},
		},
		IndexName: aws.String("GSI1"),
	})
	if err != nil {
		return nil, err
	}

	for index := range output.Items {
		item := output.Items[index]
		talent := talentFromDynamodbOutput(item)
		talentResolver := &TalentResolver{s: &talent}
		talentResolvers = append(talentResolvers, talentResolver)
	}

	return talentResolvers, nil

}

type TalentArgs struct {
	Pk           string
	Sk           string
	Name         string
	Description  string
	ProfileImage *string
}

func (r *Resolver) Talent(ctx context.Context, args TalentArgs) (*TalentResolver, error) {

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
	talent := talentFromDynamodbOutput(item)
	talentResolver := &TalentResolver{s: &talent}

	return talentResolver, nil

}

func (r *Resolver) CompanyTalents(ctx context.Context, args TalentArgs) ([]*TalentResolver, error) {

	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}
	if fmt.Sprintf("talent#%s", requestUser.CompanyId()) != args.Pk {
		log.Println("Error: user does not belong to the correct company")
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}

	keyConditionExpression := "PK = :pk"
	talentResolvers := []*TalentResolver{}
	pk := args.Pk
	output, err := r.Db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(r.TableName),
		KeyConditionExpression: aws.String(keyConditionExpression),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: &pk},
		},
	})
	if err != nil {
		return nil, err
	}

	for index := range output.Items {
		item := output.Items[index]
		talent := talentFromDynamodbOutput(item)
		talentResolver := &TalentResolver{s: &talent}
		talentResolvers = append(talentResolvers, talentResolver)
	}

	return talentResolvers, nil

}

func (r *Resolver) CreateTalent(ctx context.Context, args TalentArgs) (*TalentResolver, error) {

	log.Println("CreateTalent")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	if !requestUser.IsSellerUser() {
		log.Print("Error: requestUser is not seller")
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}

	companyId := requestUser.CompanyId()
	talentId := uuid.NewV4().String()
	pk := fmt.Sprintf("talent#%s", companyId)
	sk := fmt.Sprintf("talent_%s", talentId)
	gsi := "talent"

	var profileImageUrl *string
	// upload image if exists
	if args.ProfileImage != nil {
		profileImageUrl, _ = r.uploadProfileImage(talentId, args.ProfileImage)
	}

	talent := Talent{
		PK:              pk,
		SK:              sk,
		Name:            args.Name,
		Description:     &args.Description,
		GSI1:            &gsi,
		ProfileImageUrl: profileImageUrl,
	}

	item, err := dynamodbattribute.MarshalMap(talent)
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

	talentResolver := &TalentResolver{s: &talent}

	return talentResolver, nil

}

func (r *Resolver) UpdateTalent(ctx context.Context, args TalentArgs) (*TalentResolver, error) {

	log.Println("UpdateTalent")
	requestUser := getRequestUser(ctx)
	if requestUser == nil {
		log.Print("Error: requestUser is nil")
		return nil, errors.New(ErrorUserIsNotAuthenticated)
	}

	companyId := requestUser.CompanyId()

	if fmt.Sprintf("talent#%s", companyId) != args.Pk {
		log.Printf("Error: request user company_id: %s does not match requested company_id: %s", companyId, args.Pk)
		return nil, errors.New(ErrorUserDoesNotHaveSellerAuth)
	}

	var profileImageUrl *string
	// upload image if exists
	if args.ProfileImage != nil {
		profileImageUrl, _ = r.uploadProfileImage(args.Sk, args.ProfileImage)
	}

	updateExpression := "SET #name = :name, #description = :description"
	expressionAttributeNames := map[string]*string{
		"#name":        aws.String("Name"),
		"#description": aws.String("Description"),
	}
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":name":        {S: aws.String(args.Name)},
		":description": {S: aws.String(args.Description)},
	}

	if profileImageUrl != nil {
		updateExpression = updateExpression + ", #profileImageUrl = :profileImageUrl"
		expressionAttributeNames["#profileImageUrl"] = aws.String("ProfileImageUrl")
		expressionAttributeValues[":profileImageUrl"] = &dynamodb.AttributeValue{S: profileImageUrl}
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

	talent := talentFromDynamodbOutput(output.Attributes)

	talentResolver := &TalentResolver{s: &talent}
	return talentResolver, nil

}

type Talent struct {
	PK              string
	SK              string
	Name            string
	Description     *string
	GSI1            *string
	ProfileImageUrl *string
}

type TalentResolver struct {
	s *Talent
}

func (u *TalentResolver) PK(ctx context.Context) string {
	return u.s.PK
}

func (u *TalentResolver) SK(ctx context.Context) string {
	return u.s.SK
}

func (u *TalentResolver) Name(ctx context.Context) string {
	return u.s.Name
}

func (u *TalentResolver) Description(ctx context.Context) *string {
	return u.s.Description
}

func (u *TalentResolver) GSI1(ctx context.Context) *string {
	return u.s.GSI1
}

func (u *TalentResolver) ProfileImageUrl(ctx context.Context) *string {
	return u.s.ProfileImageUrl
}

func talentFromDynamodbOutput(item map[string]*dynamodb.AttributeValue) Talent {

	var description, gsi1, profileImageUrl *string
	if v, ok := item["Description"]; ok {
		description = v.S
	}
	if v, ok := item["GSI1"]; ok {
		gsi1 = v.S
	}
	if v, ok := item["ProfileImageUrl"]; ok {
		profileImageUrl = v.S
	}

	talent := Talent{
		PK:              *item["PK"].S,
		SK:              *item["SK"].S,
		Name:            *item["Name"].S,
		Description:     description,
		GSI1:            gsi1,
		ProfileImageUrl: profileImageUrl,
	}
	return talent
}

func (r *Resolver) uploadProfileImage(talentId string, profileImageString *string) (*string, error) {

	log.Println("uploadProfileImage")

	imgString := *profileImageString
	if baseIndex := strings.Index(imgString, "base64"); baseIndex >= 0 {
		imgString = imgString[baseIndex+7:]
	}
	log.Println(imgString)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imgString))

	m, format, err := image.Decode(reader)
	if err != nil {
		log.Fatal("Error: ", err)
		return nil, err
	}
	log.Println("-- uploadProfileImage", format)
	buf := new(bytes.Buffer)
	if format == "jpeg" {
		log.Println("Encode jpeg")
		err = jpeg.Encode(buf, m, nil)
	} else if format == "png" {
		log.Println("Encode png")
		err = png.Encode(buf, m)
	}
	log.Println("did encode image")
	if err != nil {
		log.Println("Error encoding image", err)
		return nil, err
	}

	tempFileName := fmt.Sprintf("talents/%s/profile.jpeg", talentId)
	log.Println("temp file name", tempFileName, r.BucketName)
	_, err = r.S3Client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(r.BucketName),
		Key:                  aws.String(tempFileName),
		ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(buf.Bytes()),
		ContentType:          aws.String(format),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		StorageClass:         aws.String("INTELLIGENT_TIERING"),
	})

	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	return &tempFileName, nil

}
