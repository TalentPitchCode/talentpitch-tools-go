package helpers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// CustomClaims represents the JWT claims structure
type CustomClaims struct {
	Issuer         string `json:"iss"`
	ID             string `json:"sub"` //this is an string to get an equivalent token with those PHP generated
	IssuedAt       int64  `json:"iat"`
	ExpirationTime int64  `json:"exp"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	Avatar         string `json:"avatar"`
	About          string `json:"about"`
	AboutVideo     string `json:"about_video"`
	ProfileId      uint   `json:"profile_id"`
}

func (c CustomClaims) Valid() error {
	now := time.Now().Unix()
	if c.ExpirationTime < now {
		return jwt.NewValidationError("token is expired", jwt.ValidationErrorExpired)
	}
	if c.IssuedAt > now {
		return jwt.NewValidationError("token used before issued", jwt.ValidationErrorIssuedAt)
	}
	return nil
}

func (c CustomClaims) WithValidAt(now int64) jwt.Claims {
	nowTime := time.Unix(now, 0)
	expirationTime := nowTime.Add(time.Hour * 24).Unix()
	c.ExpirationTime = expirationTime
	c.IssuedAt = now
	return &c
}

func (c CustomClaims) GetID() uint {
	id, _ := strconv.Atoi(c.ID)
	return uint(id)
}

type UserContext struct {
	ID         string
	Name       string
	Email      string
	Avatar     string
	About      string
	AboutVideo string
	ProfileId  uint
}

// CreateToken creates a JWT token with the given user context
// secretKey should be your JWT secret
// ttlSeconds is the time to live in seconds
// refreshTTL is added if refresh is true
func CreateToken(user UserContext, url string, ttlSeconds int64, secretKey []byte, refresh bool, refreshTTL int64) (string, error) {
	iat := time.Now()
	exp := iat.Add(time.Duration(ttlSeconds) * time.Second)
	if refresh {
		exp = exp.Add(time.Duration(refreshTTL) * time.Second)
	}

	claims := CustomClaims{
		Issuer:         url,
		IssuedAt:       iat.Unix(),
		ExpirationTime: exp.Unix(),
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		Avatar:         user.Avatar,
		About:          user.About,
		AboutVideo:     user.AboutVideo,
		ProfileId:      user.ProfileId,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetTokenExpiration(tokenString string, secretKey []byte) (int64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return 0, fmt.Errorf("could not parse claims")
	}

	return claims.ExpirationTime, nil
}
