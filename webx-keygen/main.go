package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	"fmt"
	"github.com/heroku/webx/keys"
	"math/big"
	"os"
)

var curveBits = flag.Int("bits", 256, "number of elliptic curve bits [224, 256, 384, 521]")

func main() {
	flag.Parse()
	curve, err := keys.ParseCurveBits(*curveBits)
	if err != nil {
		flag.Usage()
		os.Exit(0)
	}

	// random number of keys to generate
	iterationsBig, _ := rand.Int(rand.Reader, big.NewInt(200))
	iterations := int(iterationsBig.Int64()) + 500
	fmt.Printf("Generating %d keys...\n", iterations)

	var priv *ecdsa.PrivateKey
	for i := 0; i < iterations; i++ {
		priv, err = ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			fmt.Println("Error generating key:", err)
			os.Exit(1)
		}
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
