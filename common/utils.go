package common

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"time"
)

// hsin calculates the Haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance is a helper function that calculates the distance between two locations
// More at: http://en.wikipedia.org/wiki/Haversine_formula
// Returns distance in meters
func Distance(lat1, lon1, lat2, lon2 float64) float64 {

	// Convert to radians, must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180
	r = 6378100 // Earth radius in Meters

	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)
	return 2 * r * math.Asin(math.Sqrt(h))
}

func GetRequestUser(ctx context.Context) *RequestUser {
	if u := ctx.Value(RequestUserKey); u != nil {
		ru := u.(RequestUser)
		return &ru
	}
	return nil
}

type RequestUser struct {
	userGroups []string
	companyId  string
	userId     string
}

func (u *RequestUser) IsAdminUser() bool {
	for _, ug := range u.userGroups {
		if ug == "Admin" {
			return true
		}
	}
	return false
}

func (u *RequestUser) CompanyId() string {
	return u.companyId
}

func (u *RequestUser) UserId() string {
	return u.userId
}

type LogImpl struct {
	LogType  string                 `json:"logType"`
	Ts       string                 `json:"ts"`
	Msg      string                 `json:"msg"`
	Function string                 `json:"function"`
	UserId   string                 `json:"userId"`
	IsAdmin  bool                   `json:"isAdmin"`
	Err      string                 `json:"err"`
	Props    map[string]interface{} `json:"props"`
}

func LogInfo(ctx context.Context, msg, function string, props map[string]interface{}) {

	logObj := LogImpl{
		LogType:  "info",
		Ts:       time.Now().Format(time.RFC3339),
		Msg:      msg,
		Function: function,
		Props:    props,
	}
	requestUser := GetRequestUser(ctx)
	if requestUser != nil {
		logObj.UserId = requestUser.userId
		logObj.IsAdmin = requestUser.IsAdminUser()
	}

	logString, err := json.Marshal(logObj)
	if err != nil {
		LogError(ctx, msg, function, err, props)
		return
	}

	log.Print(string(logString))

}

func LogError(ctx context.Context, msg, function string, err error, props map[string]interface{}) {

	errString := ""
	if err != nil {
		errString = err.Error()
	}
	logObj := LogImpl{
		LogType:  "error",
		Ts:       time.Now().Format(time.RFC3339),
		Msg:      msg,
		Function: function,
		Err:      errString,
		Props:    props,
	}
	requestUser := GetRequestUser(ctx)
	if requestUser != nil {
		logObj.UserId = requestUser.userId
		logObj.IsAdmin = requestUser.IsAdminUser()
	}

	logString, _ := json.Marshal(logObj)
	log.Print(string(logString))

}
