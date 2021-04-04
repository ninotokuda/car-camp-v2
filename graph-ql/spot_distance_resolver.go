package main

import (
	"context"

	"github.com/ninotokuda/carcamp_v2/common"
)

type SpotDistanceResolver struct {
	spotDistance common.SpotDistance
}

func (u SpotDistanceResolver) SpotId(ctx context.Context) string {
	return u.spotDistance.OriginSpotId()
}

func (u SpotDistanceResolver) DestinationSpotId(ctx context.Context) string {
	return u.spotDistance.DestinationSpotId()
}

func (u SpotDistanceResolver) CreationTime(ctx context.Context) string {
	return u.spotDistance.CreationTime
}

func (u SpotDistanceResolver) DistanceMeters(ctx context.Context) *float64 {
	return u.spotDistance.DistanceMeters
}

func (u SpotDistanceResolver) DistanceSeconds(ctx context.Context) *float64 {
	return u.spotDistance.DistanceSeconds
}

func (u SpotDistanceResolver) DestinationName(ctx context.Context) *string {
	return u.spotDistance.DestinationName
}

func (u SpotDistanceResolver) DestinationSpotType(ctx context.Context) *string {
	return u.spotDistance.DestinationSpotType
}

func (u SpotDistanceResolver) DestinationImageUrl(ctx context.Context) *string {
	return u.spotDistance.DestinationImageUrl
}

func (u SpotDistanceResolver) DestinationDescription(ctx context.Context) *string {
	return u.spotDistance.DestinationDescription
}
