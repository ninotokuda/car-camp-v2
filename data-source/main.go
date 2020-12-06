//go:generate go-bindata schema.graphql
package main

import (
	"context"
	"errors"
	"math"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/ninotokuda/carcamp_v2/common"
	"go.uber.org/zap"
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
	s3Client       s3iface.S3API
	db             dynamodbiface.DynamoDBAPI
	mapboxClient   common.MapboxClient
	dataBucketName string
	tableName      string
}

func NewApp() *App {

	sess, _ := session.NewSession(&aws.Config{})
	svc := s3.New(sess)
	db := dynamodb.New(sess)
	bucketName := os.Getenv("S3DataBucketName")
	tableName := os.Getenv("DynamoTableName")

	mapboxClient := common.NewMapboxClient(common.MapboxConfig{
		AccessToken: "pk.eyJ1Ijoibmlub3Rva3VkYSIsImEiOiJjazl5N3g1NjUwaTJqM3FxZGoxbmh6ZXdtIn0.wGOEHFDgRW3ObcEyCMExyQ",
		DataSetId:   "ckid44kon1tbn2bsyjh5snfbk",
		BaseUrl:     "https://api.mapbox.com",
	})
	return &App{
		s3Client:       svc,
		db:             db,
		mapboxClient:   mapboxClient,
		dataBucketName: bucketName,
		tableName:      tableName,
	}
}

func (z *App) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log := ctxzap.Extract(ctx)
	log.Info("Handler")

	// load data
	csvSpots, err := z.loadCSVData(z.dataBucketName, "stations.json")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	for i := range csvSpots {
		spot := csvSpots[i]
		// upload to dynamodb
		uploadErr := common.UploadSpot(ctx, spot, z.db, z.tableName)
		if uploadErr != nil {
			log.Error("Failed to upload spot", zap.Error(uploadErr))
			continue
		}

		// upload to mapbox
		mapboxErr := z.mapboxClient.AddFeature(ctx, spot)
		if mapboxErr != nil {
			log.Error("Failed to upload Feature", zap.Error(mapboxErr))
		}
		break // test with one
	}

	// create distances
	for i := range csvSpots {
		spot := csvSpots[i]
		ghash := spot.Geohash()
		ghash4 := ghash[:4]

		nearbySpots, err := common.GetSpotsWithGeohash(ctx, ghash4, z.db, z.tableName)
		if err != nil {
			continue
		}

		// filter spots that are withing 10km
		spotsInRange := []common.Spot{}
		for j := range nearbySpots {
			ns := nearbySpots[j]
			if ns.PK == spot.PK {
				continue
			}
			if Distance(spot.Latitude, spot.Longitude, ns.Latitude, ns.Longitude) <= 10000 {
				spotsInRange = append(spotsInRange, ns)
			}
		}

		// check if distance already loaded
		existingSpotDistances, err := common.GetSpotDistances(ctx, spot, spotsInRange, z.db, z.tableName)
		if err != nil {
			continue
		}

		noDistancesSpots := []common.Spot{}
		for j := range spotsInRange {
			sr := spotsInRange[j]
			hasSpotDistance := false
			for _, esd := range existingSpotDistances {
				if sr.SpotId() == esd.DestinationSpotId() {
					hasSpotDistance = true
					break
				}
			}
			if !hasSpotDistance {
				noDistancesSpots = append(noDistancesSpots, sr)
			}
		}

		spotGroupSize := 24
		var spotGroups [][]common.Spot
		if len(noDistancesSpots)%spotGroupSize == 0 {
			spotGroups = make([][]common.Spot, len(noDistancesSpots)/spotGroupSize)
		} else {
			spotGroups = make([][]common.Spot, len(noDistancesSpots)/spotGroupSize+1)
		}

		for j := 0; j < len(spotGroups); j++ {
			end := int(math.Min(float64(len(noDistancesSpots)), float64((j+1)*spotGroupSize)))
			spotGroups[j] = noDistancesSpots[j*spotGroupSize : end]
		}
		for j := range spotGroups {
			sg := spotGroups[j]
			resp, err := z.mapboxClient.LoadDistances(ctx, spot, sg)
			if err != nil {
				continue
			}

			distanceSpots := append([]common.Spot{spot}, sg...)
			for originIndex, origin := range distanceSpots {
				for destinationIndex, destination := range distanceSpots {
					if originIndex == destinationIndex {
						continue
					}
					distanceSeconds := resp.Durations[originIndex][destinationIndex]
					distanceMeters := resp.Distances[originIndex][destinationIndex]
					spotDistance := common.NewSpotDistance(origin, destination, distanceSeconds, distanceMeters)
					err := common.AddSpotDistance(ctx, spotDistance, z.db, z.tableName)
					if err != nil {
						log.Error("Failed to add SpotDistance", zap.Error(err))
					}
				}
			}

		}
		break // test with one

	}

	return events.APIGatewayProxyResponse{
		Body:       "",
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

// hsin calculates the Haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance is a helper function that calculates the distance between two locations
// More at: http://en.wikipedia.org/wiki/Haversine_formula
// Returns distance in meters
func Distance(lat1, lon1, lat2, lon2 float64) float64 {

	// Convert to radians, must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180
	r = 6378100 // Earth radius in Meters

	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)
	return 2 * r * math.Asin(math.Sqrt(h))
}
