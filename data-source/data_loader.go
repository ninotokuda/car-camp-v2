package main

import (
	"bytes"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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
	Reatures []Feature              `json:"features"`
}

func loadCSVData(bucketName, keyName string) (StationsObject, error) {

	var stationsObject StationsObject
	sess, _ := session.NewSession(&aws.Config{})
	svc := s3.New(sess)

	rawObject, err := svc.GetObject(
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(keyName),
		})
	if err != nil {
		return stationsObject, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(rawObject.Body)

	err = json.Unmarshal(buf.Bytes(), &stationsObject)
	if err != nil {
		return stationsObject, err
	}

	return stationsObject, nil
}
