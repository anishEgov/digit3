package security

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// KeycloakClaims represents the structure of Keycloak JWT token claims
type KeycloakClaims struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	jwt.RegisteredClaims
}

// UserInfo contains extracted user information from JWT token
type UserInfo struct {
	UserID   string
	Username string
	Roles    []string
}

// ExtractUserInfoFromToken extracts user information from JWT token without signature verification
// Note: Signature verification is handled at the gateway level
func ExtractUserInfoFromToken(authHeader string) (*UserInfo, error) {
	if authHeader == "" {
		return nil, errors.New("authorization header is empty")
	}

	// Remove "Bearer " prefix
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return nil, errors.New("authorization header must start with 'Bearer '")
	}

	// Parse token without verification (gateway already verified it)
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &KeycloakClaims{})
	if err != nil {
		return nil, errors.New("failed to parse JWT token")
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok {
		return nil, errors.New("failed to extract claims from JWT token")
	}

	userInfo := &UserInfo{
		UserID:   claims.Sub,
		Username: claims.PreferredUsername,
		Roles:    claims.RealmAccess.Roles,
	}

	return userInfo, nil
}

// GetUserInfoFromTokenString is a helper function that directly accepts token string
func GetUserInfoFromTokenString(tokenString string) (*UserInfo, error) {
	return ExtractUserInfoFromToken("Bearer " + tokenString)
}
