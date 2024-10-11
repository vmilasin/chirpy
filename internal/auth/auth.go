package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Create password hash for new users on sign-up
func CreatePasswordHash(password string) ([]byte, error) {
	pw := []byte(password)

	newHash, err := bcrypt.GenerateFromPassword(pw, 12)
	if err != nil {
		return []byte{}, err
	}

	return newHash, nil
}

// Create a new access token for a user
func CreateAccessToken(userID uuid.UUID, JWTSecret []byte) (string, error) {
	now := time.Now().UTC()
	jwtExpiration := 1 * 60 * 60

	claims := &jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtExpiration) * time.Second)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(JWTSecret)
	if err != nil {
		return "", err
	}

	return signedString, nil
}

// Authorization request using an access token
func AccessTokenAuthorization(header string, JWTSecret []byte) (uuid.UUID, error) {
	var token string
	if strings.HasPrefix(header, "Bearer ") {
		token = strings.TrimPrefix(header, "Bearer ")
		token = strings.TrimSpace(token)
	} else {
		err := errors.New("invalid or missing Authorization header")
		return uuid.Nil, err
	}

	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return uuid.Nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userID := claims.Subject
	convertedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, errors.New("failed to convert userID from subject to uuid.UUID")
	}

	return convertedUserID, nil
}

// Create a new refresh token for a user
func CreateRefreshToken() (string, time.Time, error) {

	// Create a refresh token string
	randBytes := make([]byte, 32)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", time.Now(), err
	}
	hexString := hex.EncodeToString(randBytes)

	// Define token expiration timestamp
	now := time.Now().UTC()
	expirationTimestamp := now.Add(time.Duration(60*24) * time.Hour)

	return hexString, expirationTimestamp, nil
}
