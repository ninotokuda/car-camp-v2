package main

import (
	"errors"
	"fmt"
	"log"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

type AwsTokenValidator interface {
	ValidateIdToken(idToken string) (*AWSCognitoClaims, error)
}

type awsTokenValidator struct {
	publicKeySet *jwk.Set
}

type AWSCognitoClaims struct {
	SellerId      string   `json:"custom:seller_id"`
	CognitoGroups []string `json:"cognito:groups"`
	Client_ID     string   `json:"client_id"`
	Username      string   `json:"cognito:username"`
	jwt.StandardClaims
}

func NewAwsTokenValidator(publicKeysURL string) (AwsTokenValidator, error) {
	publicKeySet, err := jwk.Fetch(publicKeysURL)
	if err != nil {
		log.Printf("failed to parse key: %s", err)
		return nil, err
	}

	return &awsTokenValidator{
		publicKeySet: publicKeySet,
	}, nil
}

func (z *awsTokenValidator) ValidateIdToken(idToken string) (*AWSCognitoClaims, error) {

	claims := AWSCognitoClaims{}
	token, err := jwt.ParseWithClaims(idToken, &claims, func(t *jwt.Token) (interface{}, error) {

		// Verify if the token was signed with correct signing method
		// AWS Cognito is using RSA256 in my case
		_, ok := t.Method.(*jwt.SigningMethodRSA)

		if !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		// Get "kid" value from token header
		// "kid" is shorthand for Key ID
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid header not found")
		}

		// Check client_id attribute from the token
		claims, ok := t.Claims.(*AWSCognitoClaims)
		if !ok {
			return nil, errors.New("There is problem to get claims")
		}
		log.Printf("client_id: %v", claims.Client_ID)

		// "kid" must be present in the public keys set
		keys := z.publicKeySet.LookupKeyID(kid)
		if len(keys) == 0 {
			return nil, fmt.Errorf("key %v not found", kid)
		}

		// In our case, we are returning only one key = keys[0]
		// Return token key as []byte{string} type
		var tokenKey interface{}
		if err := keys[0].Raw(&tokenKey); err != nil {
			return nil, errors.New("failed to create token key")
		}

		return tokenKey, nil

	})

	if err != nil {
		// This place can throw expiration error
		log.Printf("token problem: %s", err)
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		log.Println("token is invalid")
		return nil, errors.New("token is invalid")
	}

	return &claims, nil

}

type mockAwsTokenValidator struct {
	ValidateIdTokenFunc func(idToken string) (*AWSCognitoClaims, error)
}

func (z *mockAwsTokenValidator) ValidateIdToken(idToken string) (*AWSCognitoClaims, error) {
	return z.ValidateIdTokenFunc(idToken)
}
