package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

type Spot struct {
	PK              string    `dynamodbav:"PK"`             // Spot#<spot_id>
	SK              string    `dynamodbav:"SK"`             // Geohash#<geohash>
	GSI1            *string   `dynamodbav:"GSI1,omitempty"` // User#<creatory_id>
	GSI2            *string   `dynamodbav:"GSI2,omitempty"` // spots
	CreationTime    string    `dynamodbav:"CreationTime"`
	SpotType        string    `dynamodbav:"SpotType"`
	Latitude        float64   `dynamodbav:"Latitude"`
	Longitude       float64   `dynamodbav:"Longitude"`
	Name            *string   `dynamodbav:"Name,omitempty"`
	Description     *string   `dynamodbav:"Description,omitempty"`
	Address         *string   `dynamodbav:"Address,omitempty"`
	Code            *string   `dynamodbav:"Code,omitempty"`
	Prefecture      *string   `dynamodbav:"Prefecture,omitempty"`
	City            *string   `dynamodbav:"City,omitempty"`
	HomePageUrls    *[]string `dynamodbav:"HomePageUrls,omitempty"`
	Tags            *[]string `dynamodbav:"Tags,omitempty"`
	DefaultImageUrl *string   `dynamodbav:"DefaultImageUrl,omitempty"`
}

func (s Spot) SpotId() string {
	return strings.TrimPrefix(s.PK, SpotPrefix)
}

func (s Spot) Geohash() string {
	return strings.TrimPrefix(s.SK, SpotPrefix)
}

type SpotDistance struct {
	PK                     string   `dynamodbav:"PK"`
	SK                     string   `dynamodbav:"SK"`
	GSI1                   *string  `dynamodbav:"GSI1,omitempty"`
	GSI2                   *string  `dynamodbav:"GSI2,omitempty"`
	CreationTime           string   `dynamodbav:"CreationTime"`
	DistanceSeconds        *float64 `dynamodbav:"DistanceSeconds,omitempty"`
	DistanceMeters         *float64 `dynamodbav:"DistanceMeters,omitempty"`
	DestinationName        *string  `dynamodbav:"DestinationName,omitempty"`
	DestinationSpotType    *string  `dynamodbav:"DestinationSpotType,omitempty"`
	DestinationImageUrl    *string  `dynamodbav:"DestinationImageUrl,omitempty"`
	DestinationDescription *string  `dynamodbav:"DestinationDescription,omitempty"`
}

func (s SpotDistance) OriginSpotId() string {
	return strings.TrimPrefix(s.PK, SpotPrefix)
}

func (s SpotDistance) DestinationSpotId() string {
	if s.GSI1 != nil {
		return strings.TrimPrefix(*s.GSI1, SpotPrefix)
	}
	return ""
}

func NewSpotDistance(origin, destination Spot, DistanceSeconds, DistanceMeters float64) SpotDistance {
	pk := fmt.Sprintf("%s%s", SpotPrefix, origin.SpotId())
	gsi1 := fmt.Sprintf("%s%s", SpotPrefix, destination.SpotId())
	creationTime := time.Now().Format(time.RFC3339)
	spotDistance := SpotDistance{
		PK:                     pk,
		SK:                     SpotDistancesQueryName,
		GSI1:                   aws.String(gsi1),
		GSI2:                   aws.String(SpotDistancesQueryName),
		CreationTime:           creationTime,
		DistanceSeconds:        aws.Float64(DistanceSeconds),
		DistanceMeters:         aws.Float64(DistanceMeters),
		DestinationName:        destination.Name,
		DestinationImageUrl:    destination.DefaultImageUrl,
		DestinationDescription: destination.Description,
	}

	return spotDistance
}
