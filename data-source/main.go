//go:generate go-bindata schema.graphql
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	"github.com/mmcloughlin/geohash"
	"github.com/ninotokuda/carcamp_v2/common"
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
		AccessToken: "sk.eyJ1Ijoibmlub3Rva3VkYSIsImEiOiJja2lkNml2Y24wNmthMnlydzV4NHY4NWZ3In0.rUAecoTts1ppwey4sDaGGA",
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

func (z *App) uploadRoadSideStations(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("uploadRoadSideStations")

	// load data
	csvSpots, err := z.loadCSVData(z.dataBucketName, "stations.json")
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
		}, err
	}

	// crate spots
	for i := range csvSpots {
		spot := csvSpots[i]

		// check if spot with geohash and type exists
		geohashSpots, err := common.GetSpotsWithGeohash(ctx, spot.Geohash(), z.db, z.tableName)
		if err != nil {
			log.Println("Failed to get geohash spots:", err.Error())
			continue
		}
		log.Println("Did fetch geohash spots", len(geohashSpots), spot.Geohash())

		hasSpot := false
		for _, gs := range geohashSpots {
			if gs.SpotType == spot.SpotType && gs.Geohash() == spot.Geohash() {
				hasSpot = true
				break
			}
		}

		if hasSpot {
			log.Println("Spot is already in db", *spot.Name)
			continue
		}

		// upload to dynamodb
		uploadErr := common.UploadSpot(ctx, spot, z.db, z.tableName)
		if uploadErr != nil {
			log.Println("Failed to upload spot:", uploadErr.Error())
			continue
		}

		// upload to mapbox
		mapboxErr := z.mapboxClient.AddFeature(ctx, spot)
		if mapboxErr != nil {
			log.Println("Failed to upload Feature:", mapboxErr.Error())
		}
	}

	// create spot distances
	for i := range csvSpots {
		spot := csvSpots[i]
		ghash := spot.Geohash()
		ghash4 := ghash[:4]

		ghash4s := geohash.Neighbors(ghash4)
		nearbySpots := []common.Spot{}
		for _, g4 := range ghash4s {
			ns, err := common.GetSpotsWithGeohash(ctx, g4, z.db, z.tableName)
			if err != nil {
				log.Println("Failed to get nearby Spots:", err.Error())
				continue
			}
			nearbySpots = append(nearbySpots, ns...)
		}

		if len(nearbySpots) == 0 {
			continue
		}

		// filter spots that are withing 10km
		spotsInRange := []common.Spot{}
		for j := range nearbySpots {
			ns := nearbySpots[j]
			if ns.PK == spot.PK {
				continue
			}
			fmt.Println("--dd")
			if common.Distance(spot.Latitude, spot.Longitude, ns.Latitude, ns.Longitude) <= 10000 {
				spotsInRange = append(spotsInRange, ns)
			}
		}

		if len(spotsInRange) == 0 {
			continue
		}

		// check if distance already loaded
		existingSpotDistances, err := common.GetSpotDistances(ctx, spot, spotsInRange, z.db, z.tableName)
		if err != nil {
			log.Println("Failed get existing spot distances:", err.Error())
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

		if len(noDistancesSpots) == 0 {
			continue
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
				log.Println("Failed to Load distances:", err.Error())
				continue
			}

			distanceSpots := append([]common.Spot{spot}, sg...)
			for originIndex, origin := range distanceSpots {
				for destinationIndex, destination := range distanceSpots {
					if originIndex != destinationIndex {
						distanceSeconds := resp.Durations[originIndex][destinationIndex]
						distanceMeters := resp.Distances[originIndex][destinationIndex]
						spotDistance := common.NewSpotDistance(origin, destination, distanceSeconds, distanceMeters)
						err := common.AddSpotDistance(ctx, spotDistance, z.db, z.tableName)
						if err != nil {
							log.Println("Failed to add SpotDistance:", err.Error())
						}
					}
				}
			}
		}
	}

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil
}

func (z *App) updateMapboxDataSet(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.Println("updateMapboxDataSet")
	// load all spots
	var lastKey string
	allSpots := []common.Spot{}
	for {
		spots, lastKey, err := common.GetAllSpots(ctx, lastKey, z.db, z.tableName)
		if err != nil {
			log.Println("Error loading all spots", err.Error())
			break
		}
		allSpots = append(allSpots, spots...)
		if lastKey == "" {
			break
		}
	}

	log.Println("Did fetch spots", len(allSpots))
	for i := range allSpots {
		spot := allSpots[i]
		// upload to mapbox
		mapboxErr := z.mapboxClient.AddFeature(ctx, spot)
		if mapboxErr != nil {
			log.Println("Failed to upload Feature:", mapboxErr.Error())
		}
	}

	log.Println("Did update mapbox features")

	// add spotid to properties\

	return events.APIGatewayProxyResponse{
		Body:       "",
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
		},
	}, nil

}

func (z *App) handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return z.updateMapboxDataSet(ctx, request)
}

func main() {
	app := NewApp()
	lambda.Start(app.handler)
}

// // hsin calculates the Haversin(θ) function
// func hsin(theta float64) float64 {
// 	return math.Pow(math.Sin(theta/2), 2)
// }

// // Distance is a helper function that calculates the distance between two locations
// // More at: http://en.wikipedia.org/wiki/Haversine_formula
// // Returns distance in meters
// func Distance(lat1, lon1, lat2, lon2 float64) float64 {

// 	// Convert to radians, must cast radius as float to multiply later
// 	var la1, lo1, la2, lo2, r float64
// 	la1 = lat1 * math.Pi / 180
// 	lo1 = lon1 * math.Pi / 180
// 	la2 = lat2 * math.Pi / 180
// 	lo2 = lon2 * math.Pi / 180
// 	r = 6378100 // Earth radius in Meters

// 	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)
// 	return 2 * r * math.Asin(math.Sqrt(h))
// }
