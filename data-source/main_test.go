package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/ninotokuda/carcamp_v2/common"
	"github.com/stretchr/testify/require"
)

func TestLoadSpots(t *testing.T) {

	testCases := []struct {
		name string
	}{
		{"load spots"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			app := &App{
				s3Client: &mockS3Client{
					GetObjectFunc: func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
						file, _ := os.Open("test_data/spots.json")
						body := ioutil.NopCloser(file)
						return &s3.GetObjectOutput{
							Body: body,
						}, nil
					},
				},
			}

			spots, err := app.loadCSVData("", "")
			require.Nil(t, err)
			require.Equal(t, 5, len(spots))
			spot1 := spots[0]
			require.Equal(t, "三笠", *spot1.Name)
			require.Equal(t, 35.067784, spot1.Latitude)
			require.Equal(t, 137.0011201, spot1.Longitude)
			require.Equal(t, "北海道", *spot1.Prefecture)
			require.Equal(t, "三笠市", *spot1.City)
			require.Equal(t, "01222", *spot1.Code)
			require.Equal(t, "北海道 三笠市", *spot1.Address)
			require.Equal(t, "spots", *spot1.GSI2)
			require.Equal(t, "RoadSideStation", spot1.SpotType)

			require.Equal(t, 3, len(spot1.HomePageUrls))
			pages := []string{*spot1.HomePageUrls[0], *spot1.HomePageUrls[1], *spot1.HomePageUrls[2]}
			p := []string{"https://www.michi-no-eki.jp/stations/view/1", "http://www.hokkaido-michinoeki.jp/michinoeki/172/", "http://www.city.mikasa.hokkaido.jp/hotnews/detail/00000036.html"}
			require.Equal(t, p, pages)
			require.Equal(t, 6, len(spot1.Tags))
			tags := []string{*spot1.Tags[0], *spot1.Tags[1], *spot1.Tags[2], *spot1.Tags[3], *spot1.Tags[4], *spot1.Tags[5]}
			ts := []string{"Atm", "BabyBed", "Restaurant", "TouristInformation", "HandicappedToilet", "Shop"}
			require.Equal(t, ts, tags)

		})
	}
}

func TestHandler(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"success"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			featureAddedCount := 0
			spotUploadedCount := 0
			spotDistanceUploadedCount := 0
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(r.URL.Path, "/datasets/") {
					featureAddedCount++
				} else if strings.Contains(r.URL.Path, "/directions-matrix/") {
					mockJson, _ := ioutil.ReadFile("test_data/distances.json")
					fmt.Fprintln(w, string(mockJson))
				}
			}))

			s3Client := &mockS3Client{
				GetObjectFunc: func(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
					file, _ := os.Open("test_data/spots.json")
					body := ioutil.NopCloser(file)
					return &s3.GetObjectOutput{
						Body: body,
					}, nil
				},
			}
			addedSpots := []map[string]*dynamodb.AttributeValue{}
			addedSpotDistances := []map[string]*dynamodb.AttributeValue{}
			dbClient := &mockDbClient{
				PutItemFunc: func(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
					if in.ConditionExpression != nil {
						addedSpots = append(addedSpots, in.Item)
						spotUploadedCount++
					} else {
						addedSpotDistances = append(addedSpotDistances, in.Item)
						spotDistanceUploadedCount++
					}
					return nil, nil
				},
				QueryFunc: func(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {

					if strings.Contains(*in.KeyConditionExpression, "#pk = :pk AND #sk IN(") {
						spotDistances := []map[string]*dynamodb.AttributeValue{}
						for i := range addedSpotDistances {
							sd := addedSpotDistances[i]
							if *sd["PK"].S == *in.ExpressionAttributeValues[":pk"].S {
								spotDistances = append(spotDistances, sd)
							}
						}
						return &dynamodb.QueryOutput{
							Items: spotDistances,
						}, nil
					}

					return &dynamodb.QueryOutput{
						Items: addedSpots,
					}, nil
				},
			}

			app := App{
				mapboxClient: &common.MapboxClientImpl{
					BaseUrl: ts.URL,
				},
				s3Client: s3Client,
				db:       dbClient,
			}

			request := events.APIGatewayProxyRequest{}
			resp, err := app.handler(context.Background(), request)
			require.Nil(t, err)
			require.Equal(t, 200, resp.StatusCode)
			require.Equal(t, 5, spotUploadedCount)
			require.Equal(t, 5, featureAddedCount)
			require.Equal(t, 6, spotDistanceUploadedCount)
		})
	}
}

type mockS3Client struct {
	s3iface.S3API
	GetObjectFunc func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

func (m *mockS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return m.GetObjectFunc(in)
}

type mockDbClient struct {
	dynamodbiface.DynamoDBAPI
	PutItemFunc func(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
	QueryFunc   func(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

func (m *mockDbClient) PutItem(in *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return m.PutItemFunc(in)
}

func (m *mockDbClient) Query(in *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return m.QueryFunc(in)
}
