package main

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

func getRequestUser(ctx context.Context) *RequestUser {
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

func logInfo(ctx context.Context, msg, function string, props map[string]interface{}) {

	logObj := LogImpl{
		LogType:  "info",
		Ts:       time.Now().Format(time.RFC3339),
		Msg:      msg,
		Function: function,
		Props:    props,
	}
	requestUser := getRequestUser(ctx)
	if requestUser != nil {
		logObj.UserId = requestUser.userId
		logObj.IsAdmin = requestUser.IsAdminUser()
	}

	logString, err := json.Marshal(logObj)
	if err != nil {
		logError(ctx, msg, function, err, props)
		return
	}

	log.Print(string(logString))

}

func logError(ctx context.Context, msg, function string, err error, props map[string]interface{}) {

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
	requestUser := getRequestUser(ctx)
	if requestUser != nil {
		logObj.UserId = requestUser.userId
		logObj.IsAdmin = requestUser.IsAdminUser()
	}

	logString, _ := json.Marshal(logObj)
	log.Print(string(logString))

}
