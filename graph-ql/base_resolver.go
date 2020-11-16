package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type QueryRequest struct {
	Query     string                 `json:"query"`
	OpName    string                 `json:"opName"`
	Variables map[string]interface{} `json:"variables"`
}

type Resolver struct {
	S3Client   s3iface.S3API
	BucketName string
	Db         dynamodbiface.DynamoDBAPI
	TableName  string
}
