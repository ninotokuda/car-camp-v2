package main

import "context"

type SpotDistanceArgs struct {
	SpotId string
}

type SpotDistance struct {
	SpotId            string
	DestinationSpotId string
	CreationTime      string
	DistanceMeters    *int32
	DistanceSeconds   *int32
}

type SpotDistanceResolver struct {
	spotDistance SpotDistance
}

func (u SpotDistanceResolver) SpotId(ctx context.Context) string {
	return u.spotDistance.SpotId
}

func (u SpotDistanceResolver) DestinationSpotId(ctx context.Context) string {
	return u.spotDistance.DestinationSpotId
}

func (u SpotDistanceResolver) CreationTime(ctx context.Context) string {
	return u.spotDistance.CreationTime
}

func (u SpotDistanceResolver) DistanceMeters(ctx context.Context) *int32 {
	return u.spotDistance.DistanceMeters
}

func (u SpotDistanceResolver) DistanceSeconds(ctx context.Context) *int32 {
	return u.spotDistance.DistanceSeconds
}
