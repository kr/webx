package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/heroku/webx/keys"
	"os"
)

func main() {
	// random number of keys to generate
	fmt.Println("Generating keys...")

	priv, err := ecdsa.GenerateKey(keys.Curve, rand.Reader)
	if err != nil {
		fmt.Println("Error generating key:", err)
		os.Exit(1)
	}

	printKey(priv)
}

func printKey(priv *ecdsa.PrivateKey) {
	fmt.Println("Private key D:\t", priv.D)
	fmt.Println("Public key X:\t", priv.PublicKey.X)
	fmt.Println("Public key Y:\t", priv.PublicKey.Y, "\n")

	fmt.Println("PrivateKey base64:")
	fmt.Println(keys.EncodePrivateKey(*priv), "\n")
	fmt.Println("PublicKey base64:")
	fmt.Println(keys.EncodePublicKey(priv.PublicKey))
}
