package main

import "context"

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

func (u *RequestUser) IsSellerUser() bool {
	for _, ug := range u.userGroups {
		if ug == "Seller" {
			return true
		}
	}
	return false
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
