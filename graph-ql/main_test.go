package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/graph-gophers/graphql-go"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {

	testCases := []struct {
		name            string
		query           string
		queryResponse   string
		getItemResponse string
		response        string
		userClaims      *AWSCognitoClaims
		errorString     string
	}{
		{
			"list companies",
			listCompaniesQuery,
			"test_data/companies.json",
			"",
			`{"data":{"listCompanies":[{"Name":"test company 1","Description":"desc"},{"Name":"test company 2","Description":"desc"}]}}`,
			adminUserClaims,
			"",
		},
		{
			"all talents",
			allTalentsQuery,
			"test_data/talents.json",
			"",
			`{"data":{"allTalents":[{"PK":"talent#company_1","SK":"talent_1","Name":"talent 1","Description":"desc"},{"PK":"talent#company_1","SK":"talent_2","Name":"talent 2","Description":"desc"}]}}`,
			adminUserClaims,
			"",
		},
		{
			"company talents",
			campanyTalentsQuery,
			"test_data/talents.json",
			"",
			`{"data":{"companyTalents":[{"PK":"talent#company_1","SK":"talent_1","Name":"talent 1","Description":"desc"},{"PK":"talent#company_1","SK":"talent_2","Name":"talent 2","Description":"desc"}]}}`,
			sellerUser1Claims,
			"",
		},
		{
			"open spots",
			openTalentSpotsQuery,
			"test_data/spots.json",
			"",
			`{"data":{"openTalentSpots":[{"Name":"test1"},{"Name":"test2"},{"Name":"test3"},{"Name":"test4"}]}}`,
			nil,
			"",
		},
		{
			"company spots",
			companySpotsQuery,
			"test_data/spots.json",
			"",
			`{"data":{"companySpots":[{"Name":"test1"},{"Name":"test2"},{"Name":"test3"},{"Name":"test4"}]}}`,
			sellerUser1Claims,
			"",
		},
		{
			"company spots wrong seller",
			companySpotsQuery,
			"test_data/spots.json",
			"",
			`{"errors":[{"message":"ErrorUserDoesNotHaveSellerAuth","path":["companySpots"]}],"data":null}`,
			sellerUser2Claims,
			"",
		},
		{
			"company spots non seller",
			companySpotsQuery,
			"test_data/spots.json",
			"",
			`{"errors":[{"message":"ErrorUserDoesNotHaveSellerAuth","path":["companySpots"]}],"data":null}`,
			user1Claims,
			"",
		},
		{
			"company spots unauthenticated",
			companySpotsQuery,
			"test_data/spots.json",
			"",
			`{"errors":[{"message":"ErrorUserIsNotAuthenticated","path":["companySpots"]}],"data":null}`,
			nil,
			"",
		},
		{
			"user spots",
			userSpotsQuery,
			"test_data/spots.json",
			"",
			`{"data":{"userSpots":[{"Name":"test1"},{"Name":"test2"},{"Name":"test3"},{"Name":"test4"}]}}`,
			user1Claims,
			"",
		},
		{
			"spot Detail",
			spotQuery,
			"",
			"test_data/spot.json",
			`{"data":{"spot":{"TalentName":"new 1","Status":"reserved"}}}`,
			user1Claims,
			"",
		},

		{
			"user Detail",
			userQuery,
			"",
			"test_data/user.json",
			`{"data":{"user":{"Nickname":"test user 111"}}}`,
			user1Claims,
			"",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			app := createTestApp(tc.queryResponse, tc.getItemResponse)
			app.awsTokenValidator = &mockAwsTokenValidator{
				ValidateIdTokenFunc: func(idToken string) (*AWSCognitoClaims, error) {
					return tc.userClaims, nil
				},
			}
			request := createTestRequest(tc.query, tc.userClaims != nil)

			resp, err := app.handler(context.Background(), request)
			require.Equal(t, tc.response, resp.Body)
			require.Nil(t, err)

		})
	}
}

func TestCreateTalent(t *testing.T) {

	testCases := []struct {
		name        string
		errorString string
		userClaims  *AWSCognitoClaims
	}{
		{
			"create talent",
			"",
			sellerUser1Claims,
		},
		{
			"non seller",
			ErrorUserDoesNotHaveSellerAuth,
			user1Claims,
		},
		{
			"notAuthenticated",
			ErrorUserIsNotAuthenticated,
			nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			didCreateItem := false
			data, _ := Asset(SchemaName)
			schemaString := string(data)
			db := &mockClientClient{
				PutItemFunc: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
					didCreateItem = true
					require.Equal(t, aws.String("talent#company_1"), input.Item["PK"].S)
					require.Contains(t, *input.Item["SK"].S, "talent_")
					require.Equal(t, aws.String("new talent name"), input.Item["Name"].S)
					require.Equal(t, aws.String("new talent desc"), input.Item["Description"].S)
					output := dynamodb.PutItemOutput{}
					return &output, nil
				},
			}
			tableName := "test_table"
			resolver := Resolver{
				Db:        db,
				TableName: tableName,
			}
			schema := graphql.MustParseSchema(schemaString, &resolver, graphql.UseStringDescriptions())

			app := &App{schema: schema}
			app.awsTokenValidator = &mockAwsTokenValidator{
				ValidateIdTokenFunc: func(idToken string) (*AWSCognitoClaims, error) {
					return tc.userClaims, nil
				},
			}

			request := createTestRequest(createTalentMutation, tc.userClaims != nil)
			resp, err := app.handler(context.Background(), request)
			if tc.errorString == "" {
				require.True(t, didCreateItem)
				require.Equal(t, "{\"data\":{\"createTalent\":{\"Name\":\"new talent name\",\"Description\":\"new talent desc\"}}}", resp.Body)
			} else {
				errorBody := fmt.Sprintf(`{"errors":[{"message":"%s","path":["createTalent"]}],"data":null}`, tc.errorString)
				require.Equal(t, errorBody, resp.Body)
				require.False(t, didCreateItem)
			}
			require.Nil(t, err)

		})
	}

}

func TestUpdateTalent(t *testing.T) {

	testCases := []struct {
		name        string
		errorString string
		userClaims  *AWSCognitoClaims
	}{
		{
			"create talent",
			"",
			sellerUser1Claims,
		},
		{
			"wrong seller",
			ErrorUserDoesNotHaveSellerAuth,
			sellerUser2Claims,
		},
		{
			"non seller",
			ErrorUserDoesNotHaveSellerAuth,
			sellerUser2Claims,
		},
		{
			"user not authenticated",
			ErrorUserIsNotAuthenticated,
			nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			didUpdateItem := false
			data, _ := Asset(SchemaName)
			schemaString := string(data)
			db := &mockClientClient{

				UpdateItemFunc: func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
					didUpdateItem = true
					require.Equal(t, aws.String("SET #name = :name, #description = :description"), input.UpdateExpression)
					return &dynamodb.UpdateItemOutput{
						Attributes: map[string]*dynamodb.AttributeValue{
							"PK":          {S: aws.String("talent#company_1")},
							"SK":          {S: aws.String("talent_1")},
							"Name":        {S: aws.String("new talent name")},
							"Description": {S: aws.String("new talent desc")},
						},
					}, nil
				},
			}
			tableName := "test_table"
			resolver := Resolver{
				Db:        db,
				TableName: tableName,
			}
			schema := graphql.MustParseSchema(schemaString, &resolver, graphql.UseStringDescriptions())

			app := &App{schema: schema}
			app.awsTokenValidator = &mockAwsTokenValidator{
				ValidateIdTokenFunc: func(idToken string) (*AWSCognitoClaims, error) {
					return tc.userClaims, nil
				},
			}

			request := createTestRequest(updateTalentMutation, tc.userClaims != nil)
			resp, err := app.handler(context.Background(), request)

			if tc.errorString == "" {
				require.True(t, didUpdateItem)
				require.Equal(t, "{\"data\":{\"updateTalent\":{\"PK\":\"talent#company_1\",\"SK\":\"talent_1\",\"Name\":\"new talent name\",\"Description\":\"new talent desc\"}}}", resp.Body)
			} else {
				errorBody := fmt.Sprintf(`{"errors":[{"message":"%s","path":["updateTalent"]}],"data":null}`, tc.errorString)
				require.Equal(t, errorBody, resp.Body)
				require.False(t, didUpdateItem)
			}
			require.Nil(t, err)
		})
	}
}

func TestCreateSpot(t *testing.T) {

	testCases := []struct {
		name        string
		userClaims  *AWSCognitoClaims
		errorString string
	}{
		{
			"create spot",
			sellerUser1Claims,
			"",
		},
		{
			"not seller",
			user1Claims,
			ErrorUserDoesNotHaveSellerAuth,
		},
		{
			"not Auth",
			nil,
			ErrorUserIsNotAuthenticated,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			didCreateItem := false
			data, _ := Asset(SchemaName)
			schemaString := string(data)
			newSpotId := ""
			db := &mockClientClient{
				PutItemFunc: func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
					didCreateItem = true
					newSpotId = *input.Item["SpotId"].S
					require.Equal(t, aws.String("spot#company_1"), input.Item["PK"].S)
					require.Contains(t, *input.Item["SK"].S, "2020/09/20T18:30:00#spot_")
					require.Contains(t, *input.Item["SpotId"].S, "spot_")
					require.Equal(t, aws.String("open"), input.Item["Status"].S)
					require.Equal(t, aws.String("spot#talent_1"), input.Item["GSI1"].S)
					require.Equal(t, aws.String("2020/09/20T18:30:00"), input.Item["Date"].S)
					require.Equal(t, aws.String("360"), input.Item["DurationSeconds"].N)
					require.Equal(t, aws.String("meet and greet"), input.Item["Name"].S)
					_, hasDescription := input.Item["Description"]
					require.False(t, hasDescription)
					output := dynamodb.PutItemOutput{}
					return &output, nil
				},
			}
			tableName := "test_table"
			resolver := Resolver{
				Db:        db,
				TableName: tableName,
			}
			schema := graphql.MustParseSchema(schemaString, &resolver, graphql.UseStringDescriptions())

			app := &App{schema: schema}
			app.awsTokenValidator = &mockAwsTokenValidator{
				ValidateIdTokenFunc: func(idToken string) (*AWSCognitoClaims, error) {
					return tc.userClaims, nil
				},
			}

			request := createTestRequest(createSpotMutation, tc.userClaims != nil)
			resp, err := app.handler(context.Background(), request)
			if tc.errorString == "" {
				require.True(t, didCreateItem)
				responseBody := fmt.Sprintf(`{"data":{"createSpot":{"PK":"spot#company_1","SK":"2020/09/20T18:30:00#%s","Date":"2020/09/20T18:30:00","DurationSeconds":360,"Name":"meet and greet","Description":null}}}`, newSpotId)
				require.Equal(t, responseBody, resp.Body)
			} else {
				errorBody := fmt.Sprintf(`{"errors":[{"message":"%s","path":["createSpot"]}],"data":null}`, tc.errorString)
				require.Equal(t, errorBody, resp.Body)
				require.False(t, didCreateItem)
			}
			require.Nil(t, err)

		})
	}

}

func TestReserveSpot(t *testing.T) {

	testCases := []struct {
		name               string
		userClaims         *AWSCognitoClaims
		err                string
		spotIsNotAvailable bool
	}{
		{"reserve spot", user1Claims, "", false},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

		})
	}

}
