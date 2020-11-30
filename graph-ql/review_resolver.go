package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type ReviewArgs struct {
	ReviewId     *string
	SpotId       *string
	UserId       *string
	LastReviewId *string
	Rating       *int32
	Message      *string
}

func (r *Resolver) Review(ctx context.Context, args ReviewArgs) (*ReviewResolver, error) {

	log.Println("Review")
	pk := fmt.Sprintf("%s%s", ReviewPrefix, *args.ReviewId)
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

	var review Review
	err = dynamodbattribute.UnmarshalMap(output.Item, &review)
	if err != nil {
		return nil, err
	}
	reviewResolver := &ReviewResolver{review: review}

	return reviewResolver, nil

}

func (r *Resolver) Reviews(ctx context.Context, args ReviewArgs) ([]*ReviewResolver, error) {

	log.Println("Reviews")

	pk := fmt.Sprintf("%s%s", SpotPrefix, *args.SpotId)
	keyConditionExpression := "#pk = :pk AND begins_with(#sk, :sk)"
	expressionAttributeValues := map[string]*dynamodb.AttributeValue{
		":pk": {S: aws.String(pk)},
		":sk": {S: aws.String("review")},
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
	})

	if err != nil {
		return nil, err
	}
	reviewResolvers := make([]*ReviewResolver, len(output.Items))
	for index := range output.Items {
		item := output.Items[index]
		var review Review
		err := dynamodbattribute.UnmarshalMap(item, &review)
		if err != nil {
			continue
		}
		reviewResolver := &ReviewResolver{review: review}
		reviewResolvers[index] = reviewResolver
	}

	return reviewResolvers, nil

}

func (r *Resolver) CreateReview(ctx context.Context, args ReviewArgs) (*ReviewResolver, error) {

	log.Println("Reviews")

	creationTime := time.Now().Format(time.RFC3339)
	pk := fmt.Sprintf("%s%s", SpotPrefix, *args.SpotId)
	sk := fmt.Sprintf("%s%s", ReviewPrefix, creationTime)
	gsi1 := fmt.Sprintf("%s%s", ReviewPrefix, *args.ReviewId)

	review := Review{
		PK:           pk,
		SK:           sk,
		CreationTime: creationTime,
		GSI1:         aws.String(gsi1),
		Rating:       args.Rating,
		Message:      args.Message,
	}
	if args.UserId != nil {
		gsi2 := fmt.Sprintf("%s%s", UserPrefix, *args.UserId)
		review.GSI2 = aws.String(gsi2)
	}

	item, err := dynamodbattribute.MarshalMap(review)

	output, err := r.Db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(r.TableName),
		Item:      item,
	})
	if err != nil {
		log.Println("Error putting item", err)
		return nil, err
	}
	log.Println("Output", output)
	reviewResolver := ReviewResolver{review: review}
	return &reviewResolver, nil

}

type Review struct {
	PK           string  `dynamodbav:"PK"`
	SK           string  `dynamodbav:"SK"`
	GSI1         *string `dynamodbav:"GSI1"`
	GSI2         *string `dynamodbav:"GSI12"`
	CreationTime string  `dynamodbav:"CreationTime"`
	Rating       *int32  `dynamodbav:"Rating"`
	Message      *string `dynamodbav:"Message"`
}

type ReviewResolver struct {
	review       Review
	baseResolver *Resolver // consider using interface instead
}

func (u ReviewResolver) ReviewId(ctx context.Context) string {
	return strings.TrimPrefix(*u.review.GSI1, ReviewPrefix)
}

func (u ReviewResolver) SpotId(ctx context.Context) string {
	return strings.TrimPrefix(u.review.PK, SpotPrefix)
}

func (u ReviewResolver) CreationTime(ctx context.Context) string {
	return u.review.CreationTime
}

func (u ReviewResolver) UserId(ctx context.Context) *string {
	if u.review.GSI2 != nil {
		return aws.String(strings.TrimPrefix(*u.review.GSI2, UserPrefix))
	}
	return nil
}

func (u ReviewResolver) User(ctx context.Context) (*UserResolver, error) {
	if userId := u.UserId(ctx); userId != nil {
		userArgs := UserArgs{UserId: *userId}
		return u.baseResolver.User(ctx, userArgs)
	}
	return nil, nil
}

func (u ReviewResolver) Rating(ctx context.Context) *int32 {
	return u.review.Rating
}

func (u ReviewResolver) Message(ctx context.Context) *string {
	return u.review.Message
}
