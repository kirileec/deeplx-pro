package translator

import (
	"log"
	"os"
	"strings"
)

var tokens []string
var hasToken bool = true

func validateToken() {
	tokenStr := os.Getenv("TOKENS")
	if tokenStr == "" {
		hasToken = false
		log.Println("No tokens provided. The service will not authenticate, which is not secure.")
		return
		//log.Fatal("No tokens provided. Please check your TOKENS environment variable.")
	}
	tokenList := strings.Split(tokenStr, ",")
	for _, token := range tokenList {
		token = strings.TrimSpace(token)
		tokens = append(tokens, token)
	}
}

func CheckToken(token string) bool {
	return stringSliceContains(tokens, token)
}
func HasToken() bool {
	return hasToken
}

func initTokens() {
	validateToken()
}
