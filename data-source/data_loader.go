package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mmcloughlin/geohash"
	"github.com/ninotokuda/carcamp_v2/common"
	uuid "github.com/satori/go.uuid"
)

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Feature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   Geometry               `json:"geometry"`
}

type StationsObject struct {
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Crs      map[string]interface{} `json:"crs"`
	Features []Feature              `json:"features"`
}

func (z *App) loadCSVData(bucketName, keyName string) ([]common.Spot, error) {

	rawObject, err := z.s3Client.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(rawObject.Body)
	var stationsObject StationsObject
	err = json.Unmarshal(buf.Bytes(), &stationsObject)
	if err != nil {
		return nil, err
	}

	spots := make([]common.Spot, len(stationsObject.Features))
	for i, f := range stationsObject.Features {
		spots[i] = FeatureToSpot(f)
	}

	return spots, nil
}

func FeatureToSpot(in Feature) common.Spot {

	uuid := uuid.NewV4().String()

	lat := in.Properties["P35_001"].(float64)
	lng := in.Properties["P35_002"].(float64)
	ghash := geohash.Encode(lat, lng)

	pk := fmt.Sprintf("%s%s", common.SpotPrefix, uuid)
	sk := fmt.Sprintf("%s%s", common.SpotPrefix, ghash)
	gsi2 := "spots"
	prefecture := in.Properties["P35_003"].(string)
	city := in.Properties["P35_004"].(string)
	code := in.Properties["P35_005"].(string)
	name := in.Properties["P35_006"].(string)
	address := fmt.Sprintf("%s %s", prefecture, city)
	spotType := "RoadSideStation"
	creationTime := time.Now().Format(time.RFC3339)

	var homePage1, homePage2, homePage3, homePage4 string
	if val, ok := in.Properties["P35_007"].(string); ok {
		homePage1 = val
	}
	if val, ok := in.Properties["P35_008"].(string); ok {
		homePage2 = val
	}
	if val, ok := in.Properties["P35_009"].(string); ok {
		homePage3 = val
	}
	if val, ok := in.Properties["P35_010"].(string); ok {
		homePage4 = val
	}
	atm := in.Properties["P35_011"].(float64)
	babyBed := in.Properties["P35_012"].(float64)
	restaurant := in.Properties["P35_013"].(float64)
	cafe := in.Properties["P35_014"].(float64)
	lightMeal := in.Properties["P35_014"].(float64)
	hotel := in.Properties["P35_015"].(float64)
	hotSpring := in.Properties["P35_016"].(float64)
	camping := in.Properties["P35_017"].(float64)
	park := in.Properties["P35_018"].(float64)
	observatory := in.Properties["P35_019"].(float64)
	museum := in.Properties["P35_020"].(float64)
	gasStand := in.Properties["P35_021"].(float64)
	evCharging := in.Properties["P35_022"].(float64)
	wifi := in.Properties["P35_023"].(float64)
	shower := in.Properties["P35_024"].(float64)
	experienceFacility := in.Properties["P35_025"].(float64)
	touristInformation := in.Properties["P35_026"].(float64)
	handicappedToilet := in.Properties["P35_027"].(float64)
	shop := in.Properties["P35_028"].(float64)

	homePages := []*string{}
	if homePage1 != "" {
		homePages = append(homePages, aws.String(homePage1))
	}

	if homePage2 != "" {
		homePages = append(homePages, aws.String(homePage2))
	}

	if homePage3 != "" {
		homePages = append(homePages, aws.String(homePage3))
	}

	if homePage4 != "" {
		homePages = append(homePages, aws.String(homePage4))
	}

	tags := []*string{}
	if atm == 1.0 {
		tags = append(tags, aws.String("Atm"))
	}
	if babyBed == 1.0 {
		tags = append(tags, aws.String("BabyBed"))
	}
	if restaurant == 1.0 {
		tags = append(tags, aws.String("Restaurant"))
	}
	if cafe == 1.0 {
		tags = append(tags, aws.String("Cafe"))
	}
	if lightMeal == 1.0 {
		tags = append(tags, aws.String("LightMeal"))
	}
	if hotel == 1.0 {
		tags = append(tags, aws.String("Hotel"))
	}
	if hotSpring == 1.0 {
		tags = append(tags, aws.String("HotSpring"))
	}
	if camping == 1.0 {
		tags = append(tags, aws.String("Camping"))
	}
	if park == 1.0 {
		tags = append(tags, aws.String("Park"))
	}
	if observatory == 1.0 {
		tags = append(tags, aws.String("Observatory"))
	}
	if museum == 1.0 {
		tags = append(tags, aws.String("Museum"))
	}
	if gasStand == 1.0 {
		tags = append(tags, aws.String("GasStand"))
	}
	if evCharging == 1.0 {
		tags = append(tags, aws.String("EvCharging"))
	}
	if wifi == 1.0 {
		tags = append(tags, aws.String("Wifi"))
	}
	if shower == 1.0 {
		tags = append(tags, aws.String("Shower"))
	}
	if experienceFacility == 1.0 {
		tags = append(tags, aws.String("ExperienceFacility"))
	}
	if touristInformation == 1.0 {
		tags = append(tags, aws.String("TouristInformation"))
	}
	if handicappedToilet == 1.0 {
		tags = append(tags, aws.String("HandicappedToilet"))
	}
	if shop == 1.0 {
		tags = append(tags, aws.String("Shop"))
	}

	return common.Spot{
		PK:           pk,
		SK:           sk,
		GSI2:         aws.String(gsi2),
		Latitude:     lat,
		Longitude:    lng,
		Prefecture:   aws.String(prefecture),
		City:         aws.String(city),
		Name:         aws.String(name),
		Code:         aws.String(code),
		HomePageUrls: homePages,
		Tags:         tags,
		Address:      aws.String(address),
		SpotType:     spotType,
		CreationTime: creationTime,
	}
}
