package config

import "net/http"

func GetAccessTokenFromContext(r *http.Request) string {
	if value := r.Context().Value("accessToken"); value != nil {
		if token, ok := value.(string); ok {
			return token
		}
	}
	return ""
}
