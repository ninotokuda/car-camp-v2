package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type UserArgs struct {
	UserId string
}

func (r *Resolver) User(ctx context.Context, args UserArgs) (*UserResolver, error) {

	log.Println("User")
	pk := fmt.Sprintf("%s%s", UserPrefix, args.UserId)
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

	var user User
	err = dynamodbattribute.UnmarshalMap(output.Item, &user)
	if err != nil {
		return nil, err
	}
	userResolver := &UserResolver{user: user}
	return userResolver, nil

}

type User struct {
	PK           string  `dynamodbav:"PK"`
	SK           string  `dynamodbav:"SK"`
	GSI1         *string `dynamodbav:"GSI1"`
	GSI2         *string `dynamodbav:"GSI12"`
	CreationTime string  `dynamodbav:"CreationTime"`
	Nickname     *string `dynamodbav:"CreationTime"`
}

type UserResolver struct {
	user         User
	baseResolver *Resolver // consider using interface instead
}

func (u UserResolver) UserId(ctx context.Context) string {
	return strings.TrimPrefix(u.user.PK, UserPrefix)
}

func (u UserResolver) Nickname(ctx context.Context) *string {
	return u.user.Nickname
}

func (u UserResolver) CreationTime(ctx context.Context) string {
	return u.user.CreationTime
}

func (u UserResolver) Reviews(ctx context.Context) (*[]*ReviewResolver, error) {
	userId := u.UserId(ctx)
	reviewArgs := ReviewArgs{UserId: aws.String(userId)}
	resolvers, err := u.baseResolver.Reviews(ctx, reviewArgs)
	return &resolvers, err
}

func (u UserResolver) CreatedSpots(ctx context.Context) (*[]*SpotResolver, error) {
	userId := u.UserId(ctx)
	spotArgs := SpotArgs{CreatorId: userId}
	resolvers, err := u.baseResolver.SpotsByCreator(ctx, spotArgs)
	return &resolvers, err
}
