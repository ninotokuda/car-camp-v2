package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type MapboxClient interface {
	AddFeature(ctx context.Context, spot Spot) error
	LoadDistances(ctx context.Context, origin Spot, destinations []Spot) (LoadDistancesResponse, error)
}

type MapboxClientImpl struct {
	AccessToken string
	DataSetId   string
	BaseUrl     string
	Client      http.Client
}

type MapboxConfig struct {
	AccessToken string
	DataSetId   string
	BaseUrl     string
}

func NewMapboxClient(config MapboxConfig) MapboxClient {

	httpClient := http.Client{}
	return &MapboxClientImpl{
		AccessToken: config.AccessToken,
		DataSetId:   config.DataSetId,
		BaseUrl:     config.BaseUrl,
		Client:      httpClient,
	}
}

type MapboxGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type AddFeatureRequest struct {
	Id         string            `json:"id"`
	Type       string            `json:"type"`
	Geometry   MapboxGeometry    `json:"geometry"`
	Properties map[string]string `json:"properties"`
}

func (z *MapboxClientImpl) AddFeature(ctx context.Context, spot Spot) error {

	log.Println("AddFeature")
	log.Println("FeatureId: ", spot.SpotId())
	spotId := spot.SpotId()
	featureRequest := AddFeatureRequest{
		Id:   spotId,
		Type: "Feature",
		Geometry: MapboxGeometry{
			Type:        "Point",
			Coordinates: []float64{spot.Longitude, spot.Latitude},
		},
		Properties: map[string]string{
			"Name":     *spot.Name,
			"SpotType": spot.SpotType,
			"SpotId":   spot.SpotId(),
		},
	}

	featureRequestJson, err := json.Marshal(featureRequest)
	if err != nil {
		return err
	}

	requestUrl := fmt.Sprintf("%s/datasets/v1/ninotokuda/%s/features/%s?access_token=%s", z.BaseUrl, z.DataSetId, spotId, z.AccessToken)
	request, err := http.NewRequest("PUT", requestUrl, bytes.NewReader(featureRequestJson))
	if err != nil {
		return err
	}
	_, err = z.executeRequest(ctx, request)

	return err
}

type LoadDistancesResponse struct {
	Code      string      `json:"code"`
	Durations [][]float64 `json:"durations"`
	Distances [][]float64 `json:"distances"`
}

func (z *MapboxClientImpl) LoadDistances(ctx context.Context, origin Spot, destinations []Spot) (LoadDistancesResponse, error) {

	var resp LoadDistancesResponse
	coordinates := []string{fmt.Sprintf("%f,%f", origin.Longitude, origin.Latitude)}
	for _, d := range destinations {
		coordinates = append(coordinates, fmt.Sprintf("%f,%f", d.Longitude, d.Latitude))
	}
	coordinatesStr := strings.Join(coordinates, ";")
	drivingProfile := "mapbox/driving"
	requestUrl := fmt.Sprintf("%s/directions-matrix/v1/%s/%s?annotations=duration,distance&access_token=%s", z.BaseUrl, drivingProfile, coordinatesStr, z.AccessToken)
	request, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return resp, err
	}
	response, err := z.executeRequest(ctx, request)
	if err != nil {
		return resp, err
	}

	err = json.Unmarshal(response, &resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (z *MapboxClientImpl) executeRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	log.Println("executeRequest")
	req.Header.Add("Content-Type", "application/json")
	log.Println("request url", req.URL)

	response, err := z.Client.Do(req.WithContext(ctx))
	if err != nil {
		log.Println("Error making request", err.Error())
		return nil, err
	}
	defer response.Body.Close()
	var bytes []byte
	bytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Error reading Response body:", err)
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		log.Println("Error in status code", response.StatusCode, string(bytes))
		return nil, errors.New("Non success status code")
	}
	log.Println("response body", string(bytes))
	return bytes, nil
}
