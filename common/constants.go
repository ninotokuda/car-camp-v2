package common

const (
	SchemaName = "schema.graphql"

	// keys
	RequestUserKey = "request_user"

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

	// prefixes
	SpotPrefix    = "Spot#"
	UserPrefix    = "User#"
	ReviewPrefix  = "Review#"
	GeohashPrefix = "Geohash#"

	// queryNames
	SpotQueryName          = "spots"
	SpotDistancesQueryName = "SpotDistances"

	// keys
	PKKey           = "PK"
	SKKey           = "SK"
	GSI1Key         = "GSI1"
	GSI2Key         = "GSI2"
	SpotTypeKey     = "SpotType"
	LatitudeKey     = "Latitude"
	LongitudeKey    = "Longitude"
	CreatorIdKey    = "CreatorId"
	CreationTimeKey = "CreationTime"
	NameKey         = "Name"
	DescriptionKey  = "Description"
	AddressKey      = "Address"
	CodeKey         = "Code"
	PrefectureKey   = "Prefecture"
	CityKey         = "City"
	HomePageUrlsKey = "HomePageUrls"
	TagsKey         = "Tags"
	MessageKey      = "Message"
	RatingKey       = "Rating"
)
