package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/JathinShyam/NexusLink/internal/auth"
)

func main() {
	userID := flag.String("user", "dev-user", "user id to embed in the token")
	secret := flag.String("secret", os.Getenv("JWT_SECRET"), "JWT signing secret")
	ttlHours := flag.Int("ttl-hours", 24, "token lifetime in hours")
	flag.Parse()

	if *secret == "" {
		*secret = "dev-secret-change-me"
	}

	jwtService := auth.NewJWT(*secret, time.Duration(*ttlHours)*time.Hour)
	token, err := jwtService.Issue(*userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "issue token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}
