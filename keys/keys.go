package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"math/big"
)

func EncodePublicKey(pub ecdsa.PublicKey) string {
	return base64.URLEncoding.EncodeToString(append(pub.X.Bytes(), pub.Y.Bytes()...))
}

func DecodePublicKey(xy64 string, curve elliptic.Curve) (*ecdsa.PublicKey, error) {
	xy, err := base64.URLEncoding.DecodeString(xy64)
	if err != nil {
		return nil, err
	}

	x := big.NewInt(0).SetBytes(xy[:len(xy)/2])
	y := big.NewInt(0).SetBytes(xy[len(xy)/2:])
	return &ecdsa.PublicKey{curve, x, y}, nil
}

func EncodePrivateKey(priv ecdsa.PrivateKey) string {
	return base64.URLEncoding.EncodeToString(priv.D.Bytes())
}

func DecodePrivateKey(b64 string, curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
	z, err := base64.URLEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{Curve: curve},
		D:         big.NewInt(0).SetBytes(z),
	}, nil
}

func EncodeSignature(r, s *big.Int) string {
	return base64.URLEncoding.EncodeToString(append(r.Bytes(), s.Bytes()...))
}

func DecodeSignature(signature64 string) (r, s *big.Int, err error) {
	rs, err := base64.URLEncoding.DecodeString(signature64)
	if err != nil {
		return
	}

	r = big.NewInt(0).SetBytes(rs[:len(rs)/2])
	s = big.NewInt(0).SetBytes(rs[len(rs)/2:])
	return
}

func SignUser(user string, key *ecdsa.PrivateKey) (auth string, err error) {
	hs := sha512.New().Sum([]byte(user))

	r, s, err := ecdsa.Sign(rand.Reader, key, hs)
	if err != nil {
		fmt.Errorf("%s error signing: %s", hs, err)
	}

	return EncodeSignature(r, s), nil
}

func VerifySignedUser(user, signature string, pub *ecdsa.PublicKey) (bool, error) {
	r, s, err := DecodeSignature(signature)
	if err != nil {
		return false, err
	}

	return ecdsa.Verify(pub, sha512.New().Sum([]byte(user)), r, s), nil
}
