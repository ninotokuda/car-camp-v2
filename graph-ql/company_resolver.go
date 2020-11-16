package main

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func (r *Resolver) ListCompanies(ctx context.Context) ([]*CompanyResolver, error) {

	keyConditionExpression := "PK = :pk"
	companyResolvers := []*CompanyResolver{}
	pk := "company"
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
		company := Company{
			PK:          *item["PK"].S,
			SK:          *item["SK"].S,
			Name:        *item["Name"].S,
			Description: item["Description"].S,
		}
		companyResolver := &CompanyResolver{s: &company}
		companyResolvers = append(companyResolvers, companyResolver)
	}

	return companyResolvers, nil

}

type Company struct {
	PK          string
	SK          string
	CompanyId   string
	Name        string
	Description *string
}

type CompanyResolver struct {
	s *Company
}

func (u *CompanyResolver) PK(ctx context.Context) string {
	return u.s.PK
}

func (u *CompanyResolver) SK(ctx context.Context) string {
	return u.s.SK
}

func (u *CompanyResolver) CompanyId(ctx context.Context) string {
	return u.s.CompanyId
}

func (u *CompanyResolver) Name(ctx context.Context) string {
	return u.s.Name
}

func (u *CompanyResolver) Description(ctx context.Context) *string {
	return u.s.Description
}
