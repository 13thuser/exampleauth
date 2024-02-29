package main

import "os"

var JWT_SECRET_KEY = getSecretKey()

// Read the secret key from the environment variable otherwise use the default value
func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "my-secret-key"
	}
	return secretKey
}
