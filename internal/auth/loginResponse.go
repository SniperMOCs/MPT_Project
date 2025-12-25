package auth

type loginResponse struct {
	Token string `json:"access_token"`
}
