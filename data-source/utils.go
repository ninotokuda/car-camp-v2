package main

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/ninotokuda/carcamp_v2/common"
)

func createSpotDistances(ctx context.Context, origin common.Spot, destinations []common.Spot, db dynamodb.DynamoDB, tableName string) error {

	return nil
}
