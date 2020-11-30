package main

import "context"

type SpotImageArgs struct {
	SpotId string
}

type SpotImage struct {
	SpotImageId  string
	SpotId       string
	ImageUrl     string
	UserId       *string
	CreationTime *string
}

type SpotImageResolver struct {
	spotImage SpotImage
}

func (u SpotImageResolver) SpotImageId(ctx context.Context) string {
	return u.spotImage.SpotImageId
}

func (u SpotImageResolver) SpotId(ctx context.Context) string {
	return u.spotImage.SpotId
}

func (u SpotImageResolver) ImageUrl(ctx context.Context) string {
	return u.spotImage.ImageUrl
}

func (u SpotImageResolver) UserId(ctx context.Context) *string {
	return u.spotImage.UserId
}

func (u SpotImageResolver) CreationTime(ctx context.Context) *string {
	return u.spotImage.CreationTime
}
