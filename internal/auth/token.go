package auth

type TokenManager interface {
	Generate(userID string) (string, error)
	Parse(tokenString string) (string, error)
}
