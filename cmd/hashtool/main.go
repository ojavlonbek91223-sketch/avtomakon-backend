// Kichik utility: Argon2id parol heshini generatsiya qiladi
// Foydalanish: go run ./cmd/hashtool "parol"
package main

import (
	"fmt"
	"os"

	"github.com/avtomakon/backend/internal/pkg/hash"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Foydalanish: go run ./cmd/hashtool \"<parol>\"")
		os.Exit(1)
	}

	password := os.Args[1]
	hashed, err := hash.HashPassword(password)
	if err != nil {
		fmt.Printf("Xato: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hashed)
}
