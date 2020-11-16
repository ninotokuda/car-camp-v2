//go:generate go-bindata schema.graphql
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/graph-gophers/graphql-go"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

type App struct {
	schema            *graphql.Schema
	awsTokenValidator AwsTokenValidator
}

func NewApp() *App {

	data, err := Asset(SchemaName)
	if err != nil {
		panic(err)
	}
	schemaString := string(data)

	mySession := session.Must(session.NewSession())
	db := dynamodb.New(mySession)
	tableName := os.Getenv(TableNameEvn)
	s3Client := s3.New(mySession)
	bucketName := os.Getenv(BucketNameEnv)
	resolver := Resolver{
		S3Client:   s3Client,
		BucketName: bucketName,
		Db:         db,
		TableName:  tableName,
	}
	schema := graphql.MustParseSchema(schemaString, &resolver, graphql.UseStringDescriptions())

	publicKeysURL := "https://cognito-idp.ap-northeast-1.amazonaws.com/ap-northeast-1_IkvtTA79k/.well-known/jwks.json"
	awsTokenValidator, err := NewAwsTokenValidator(publicKeysURL)
	if err != nil {
		panic(err)
	}

	return &App{
		schema:            schema,
		awsTokenValidator: awsTokenValidator,
	}
}

func (z *App) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// valid idToken and create user if Authorization header is set
	if idToken, ok := request.Headers["Authorization"]; ok {
		// strip Bearer
		if strings.Contains(idToken, "Bearer ") {
			idToken = strings.Replace(idToken, "Bearer ", "", 1)
		}

		// validate claims
		claims, err := z.awsTokenValidator.ValidateIdToken(idToken)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: 403,
				Headers: map[string]string{
					"Access-Control-Allow-Origin": "*",
				},
			}, err
		}

		// create user from claims
		log.Println("--- claims", claims)
		requestUser := RequestUser{claims.CognitoGroups, claims.SellerId, claims.Username}
		log.Println("--", requestUser)
		ctx = context.WithValue(ctx, RequestUserKey, requestUser)

	}

	var queryRequest QueryRequest
	marshalErr := json.Unmarshal([]byte(request.Body), &queryRequest)
	if marshalErr != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 403,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
			},
		}, marshalErr
	}

	resp := z.schema.Exec(ctx, queryRequest.Query, queryRequest.OpName, queryRequest.Variables)
	rJSON, _ := json.Marshal(resp)
	return events.APIGatewayProxyResponse{
		Body:       string(rJSON),
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func main() {
	app := NewApp()
	lambda.Start(app.handler)
}
