package main

const (
	SchemaName = "schema.graphql"

	// keys
	RequestUserKey = "request_user"

	// db keys
	pkKey          = "PK"
	skKey          = "SK"
	gsi1Key        = "GSI1"
	gsi2Key        = "GSI2"
	nameKey        = "Name"
	DescriptionKey = "Description"

	// env vars
	TableNameEvn  = "DynamoTableName"
	BucketNameEnv = "S3BucketName"

	// spot statuses
	SpotStatusOpen     = "open"
	SpotStatusReserved = "reserved"

	// errors
	ErrorUserIsNotAuthenticated    = "ErrorUserIsNotAuthenticated"
	ErrorUserDoesNotHaveSellerAuth = "ErrorUserDoesNotHaveSellerAuth"
	ErrorSpotIsAlreadyReserved     = "ErrorSpotIsAlreadyReserved"
)
