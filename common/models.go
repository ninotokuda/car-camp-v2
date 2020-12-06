package common

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

type Spot struct {
	PK           string    `dynamodbav:"PK"`
	SK           string    `dynamodbav:"SK"`
	GSI1         *string   `dynamodbav:"GSI1"`
	GSI2         *string   `dynamodbav:"GSI12"`
	CreationTime string    `dynamodbav:"CreationTime"`
	SpotType     string    `dynamodbav:"SpotType"`
	Latitude     float64   `dynamodbav:"Latitude"`
	Longitude    float64   `dynamodbav:"Longitude"`
	Name         *string   `dynamodbav:"Name"`
	Description  *string   `dynamodbav:"Description"`
	Address      *string   `dynamodbav:"Address"`
	Code         *string   `dynamodbav:"Code"`
	Prefecture   *string   `dynamodbav:"Prefecture"`
	City         *string   `dynamodbav:"City"`
	HomePageUrls []*string `dynamodbav:"HomePageUrls"`
	Tags         []*string `dynamodbav:"Tags"`
}

func (s Spot) SpotId() string {
	return strings.TrimPrefix(s.PK, SpotPrefix)
}

func (s Spot) Geohash() string {
	return strings.TrimPrefix(s.SK, SpotPrefix)
}

type SpotDistance struct {
	PK              string  `dynamodbav:"PK"`
	SK              string  `dynamodbav:"SK"`
	GSI1            *string `dynamodbav:"GSI1"`
	DistanceSeconds float64 `dynamodbav:"DistanceSeconds"`
	DistanceMeters  float64 `dynamodbav:"DistanceMeters"`
}

func (s SpotDistance) DestinationSpotId() string {
	return strings.TrimPrefix(s.SK, SpotPrefix)
}

func NewSpotDistance(origin, destination Spot, DistanceSeconds, DistanceMeters float64) SpotDistance {
	pk := fmt.Sprintf("%s%s", SpotPrefix, origin.SpotId())
	sk := fmt.Sprintf("%s%s", SpotPrefix, destination.SpotId())
	gsi1 := fmt.Sprintf("%s%s", SpotPrefix, destination.SpotId())
	return SpotDistance{
		PK:              pk,
		SK:              sk,
		GSI1:            aws.String(gsi1),
		DistanceSeconds: DistanceSeconds,
		DistanceMeters:  DistanceMeters,
	}
}
