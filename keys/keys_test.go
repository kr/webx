package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"reflect"
	"testing"
)

func TestEncodeDecodePublicKey(t *testing.T) {
	xstr := "91902829668862169015983137365766418463234730702165833401852203800663795127946"
	ystr := "71706737214393833266931624628724922581229333163681892523661102756208006988628"
	expected := "yy8ogru0TNzU4dBEPoNopqlSfYfQ8nBp5kel9HTvVoqeiJMzNqwDFNGmq8Qisy9kcV622G4LzfnT2ylR9s0DVA=="
	x, _ := big.NewInt(0).SetString(xstr, 10)
	y, _ := big.NewInt(0).SetString(ystr, 10)

	k := ecdsa.PublicKey{elliptic.P256(), x, y}
	enc := EncodePublicKey(k)
	if enc != expected {
		t.Errorf("expected encoding %q got %q", expected, enc)
	}

	kdec, err := DecodePublicKey(enc, elliptic.P256())
	if err != nil {
		t.Errorf("error decoding: %s", err)
	} else if !reflect.DeepEqual(*kdec, k) {
		t.Errorf("expected decoded key %v, got %v", k, *kdec)
	}
}

func TestEncodeDecodePrivateKey(t *testing.T) {
	dstr := "89313698214387454822521149997320673442561542575811640358557629057740883524937"
	expected := "xXXDA5MP4k-ZqFxyyiaGEeS5zR7tkFkZbG1mrnTmfUk="
	d, _ := big.NewInt(0).SetString(dstr, 10)

	k := ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256()}, D: d}
	enc := EncodePrivateKey(k)
	if enc != expected {
		t.Errorf("expected %q got %q", expected, enc)
	}
	
	kdec, err := DecodePrivateKey(enc, elliptic.P256())
	if err != nil {
		t.Errorf("error decoding: %s", err)
	} else if !reflect.DeepEqual(*kdec, k) {
		t.Errorf("expected decoded key %v, got %v", k, *kdec)
	}
}

func TestEncodeDecodeSignature(t *testing.T) {
	rstr := "110216375824334377012853932770766869563602422959681893306863750147455361675584"
	sstr := "25274077453555616087967264826031329470496583147881004563517787762998651048412"
	expected := "86xCDyQW7XQ4eG5iO1QGrkEHKAzH1p61iBz2Wcha8UA34J7zb9FR6LBhRaPmaf5NkqT6KIGvcBq1Cq3ZWJ7R3A=="

	r, _ := big.NewInt(0).SetString(rstr, 10)
	s, _ := big.NewInt(0).SetString(sstr, 10)

	res := EncodeSignature(r, s)
	if res != expected {
		t.Errorf("expected %s, got %s", expected, res)
	}

	r2, s2, err := DecodeSignature(res)
	if err != nil {
		t.Fatalf("error decoding signature: %s", err)
	}
	if r.Cmp(r2) != 0 {
		t.Errorf("expected r=%s, got %s", r, r2)
	}
	if s.Cmp(s2) != 0 {
		t.Errorf("expected s=%s, got %s", s, s2)
	}
}

func TestSignUser(t *testing.T) {
	key64 := "4d2028ad91408478bdfb44cfb18d0de7316290e49471ffe58691a5947d12866c"
	kdec, err := DecodePrivateKey(key64, elliptic.P256())
	if err != nil {
		t.Errorf("error decoding: %s", err)
	}

	_, err = SignUser("omgwtfbbq", kdec)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestVerifySignedUser(t *testing.T) {
	curve := elliptic.P256()
	priv := genTestKey(t, curve)
	user := "omgwtfbbq"

	auth, err := SignUser(user, priv)
	if err != nil {
		t.Fatalf("error signing user: %s", err)
	}

	success, err := VerifySignedUser(user, auth, &priv.PublicKey)
	if err != nil {
		t.Fatalf("error verifying user: %s", err)
	}
	if !success {
		t.Fatalf("verification failed")
	}
}

func genTestKey(t *testing.T, curve elliptic.Curve) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		t.Fatalf("error generating test key: %s", err)
	}
	return k
}
