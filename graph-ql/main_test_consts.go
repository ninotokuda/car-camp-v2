package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/graph-gophers/graphql-go"
)

const (
	// TestHandler
	spotQuery = `{
		"query":"query Spot($spotId: String!){spot(spotId: $spotId){Name\nDescription\nReviews{\nReviewId\nRating\nMessage\n}}}",
		"variables": {"spotId":"spot1"}
	}`

	allTalentsQuery = `{
		"query":"query AllTalents{allTalents(){PK\nSK\nName\nDescription}}"
	}`

	campanyTalentsQuery = `{
		"query":"query CompanyTalents($pk: String!){companyTalents(pk: $pk){PK\nSK\nName\nDescription}}",
		"variables": {"pk":"talent#company_1"}
	}`

	openTalentSpotsQuery = `{
		"query":"query OpenTalentSpots($talentId: String!){openTalentSpots(talentId: $talentId){Name}}",
		"variables": {"talentId":"talent_1"}
	}`

	companySpotsQuery = `{
		"query":"query CompanySpots($companyId: String!){companySpots(companyId: $companyId){Name}}",
		"variables": {"companyId":"company_1"}
	}`

	userSpotsQuery = `{
		"query":"query UserSpots($userId: String!){userSpots(userId: $userId){Name}}",
		"variables": {"userId":"user_1"}
	}`

	createTalentMutation = `{
		"query" : "mutation CreateTalent($name: String!, $description: String!){createTalent(name: $name, description: $description){Name\nDescription}}",
		"variables": {"name": "new talent name","description": "new talent desc"}
	}`

	updateTalentMutation = `{
		"query" : "mutation UpdateTalent($pk: String!, $sk: String!, $name: String!, $description: String!){updateTalent(pk: $pk, sk: $sk, name: $name, description: $description){PK\nSK\nName\nDescription}}",
		"variables": {"pk": "talent#company_1", "sk":"talent_1", "name": "new talent name","description": "new talent desc"}
	}`

	createSpotMutation = `{
		"query" : "mutation CreateSpot($talentId: String!, $companyId: String!, $date: String!, $durationSeconds: Int!, $priceYen: Int!, $name: String, $description: String){createSpot(talentId: $talentId, companyId: $companyId, date: $date, durationSeconds: $durationSeconds, priceYen: $priceYen, name: $name, description: $description){PK\nSK\nDate\nDurationSeconds\nName\nDescription}}",
		"variables": {"talentId":"talent_1", "companyId":"company_1", "date":"2020/09/20T18:30:00", "durationSeconds":360, "priceYen":500, "name":"meet and greet"}
	}`

	reserveSpotMutation = `{
		"query" : "mutation ReserveSpot($pk: String!, $sk: String!, $userId: String!, $userName: String!){reserveSpot(pk: $pk, sk: $sk, userId: $userId, userName: $userName){PK\nSK\nDate\nDurationSeconds\nName\nDescription}}",
		"variables": {"pk":"spot#company_1", "sk":"open", "date":"2020/09/20T18:30:00", "durationSeconds":360, "priceYen":500, "name":"meet and greet"}
	}`

	userQuery = `{
		"query":"query User($sk: String!){user(sk: $sk){Nickname}}",
		"variables": {"sk":"user_e76fff27-ffe8-4317-a62c-5ba167f084da"}
	}`
)

var (
	// user1             = RequestUser{[]string{}, "", "user_1"}
	// sellerUser1       = RequestUser{[]string{"Seller"}, "company_1", "user_2"}
	// sellerUser2       = RequestUser{[]string{"Seller"}, "company_2", "user_3"}
	// adminUser         = RequestUser{[]string{"Admin"}, "", "user_4"}
	user1Claims = &AWSCognitoClaims{
		Username: "user_1",
	}
	sellerUser1Claims = &AWSCognitoClaims{
		SellerId:      "company_1",
		CognitoGroups: []string{"Seller"},
		Username:      "user_2",
	}
	sellerUser2Claims = &AWSCognitoClaims{
		SellerId:      "company_2",
		CognitoGroups: []string{"Seller"},
		Username:      "user_3",
	}
	adminUserClaims = &AWSCognitoClaims{
		CognitoGroups: []string{"Admin"},
		Username:      "user_4",
	}
)

func createTestRequest(query string, isAuthenticated bool) events.APIGatewayProxyRequest {

	event := events.APIGatewayProxyRequest{
		Body: query,
	}
	if isAuthenticated {
		event.Headers = map[string]string{
			"Authorization": "testIdToken",
		}
	}
	return event
}

type mockClientClient struct {
	dynamodbiface.DynamoDBAPI
	QueryFunc      func(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	PutItemFunc    func(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	UpdateItemFunc func(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error)
	GetItemFunc    func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
}

func (m *mockClientClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return m.PutItemFunc(input)
}

func (m *mockClientClient) Query(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return m.QueryFunc(input)
}

func (m *mockClientClient) UpdateItem(input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemFunc(input)
}

func (m *mockClientClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.GetItemFunc(input)
}

func createTestApp(queryResponsePath, getItemResponsePath string) *App {

	data, _ := Asset("schema.graphql")
	schemaString := string(data)
	db := &mockClientClient{
		QueryFunc: func(input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
			if queryResponsePath != "" {
				output := dynamodb.QueryOutput{}
				mockJson, _ := ioutil.ReadFile(queryResponsePath)
				json.Unmarshal(mockJson, &output)
				return &output, nil
			}
			return nil, nil
		},
		GetItemFunc: func(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
			if getItemResponsePath != "" {
				output := dynamodb.GetItemOutput{}
				mockJson, _ := ioutil.ReadFile(getItemResponsePath)
				json.Unmarshal(mockJson, &output)
				return &output, nil
			}
			return nil, nil
		},
	}
	tableName := "test_table"
	resolver := Resolver{
		Db:        db,
		TableName: tableName,
	}
	schema := graphql.MustParseSchema(schemaString, &resolver, graphql.UseStringDescriptions())

	return &App{
		schema: schema,
	}

}
