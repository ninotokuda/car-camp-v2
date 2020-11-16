package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type UserArgs struct {
	Pk string
	Sk string
}

func (r *Resolver) User(ctx context.Context, args UserArgs) (*UserResolver, error) {

	log.Println("User")
	output, err := r.Db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(r.TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("user"),
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
	user := userFromDynamoDbOutput(item)
	userResolver := &UserResolver{s: &user}

	return userResolver, nil

}

type User struct {
	PK       string
	SK       string
	Nickname *string
}

type UserResolver struct {
	s *User
}

func (u *UserResolver) PK(ctx context.Context) string {
	return u.s.PK
}

func (u *UserResolver) SK(ctx context.Context) string {
	return u.s.SK
}

func (u *UserResolver) Nickname(ctx context.Context) *string {
	return u.s.Nickname
}

func userFromDynamoDbOutput(item map[string]*dynamodb.AttributeValue) User {

	var nickname *string
	if v, ok := item["Nickname"]; ok {
		nickname = v.S
	}

	user := User{
		PK:       *item["PK"].S,
		SK:       *item["SK"].S,
		Nickname: nickname,
	}
	return user
}
