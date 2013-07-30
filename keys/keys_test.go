package keys

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"reflect"
	"testing"
)

func TestEncodeDecodePublicKey(t *testing.T) {
	xstr := "3752433171554439267614164748152619388643469309054567599316549718761947465529723419288751913782382747582304433855580769125882294263944376115280908689991506544"
	ystr := "4152006475681996459286906069233895865610009201015117130302757857770883807111947238111213586760517239268796365966145179647128950410979775566153140254442410978"
	expected := "ARfeh0abwJUHTINEQ0o7_NxnG8gda3s5AywydxovC3i148GvmCEeRIVNRCuMIuCcB71dEAEngOFc1n_fCefOL5pwATWruOcIlyO65lIqCkAoK3SmpIhSbVnzPsdKkPh2Rwwt61uJYSvtY3JW7gdop4vG3g8-FxETcMMEZGH57vzTUlPi"
	x, _ := big.NewInt(0).SetString(xstr, 10)
	y, _ := big.NewInt(0).SetString(ystr, 10)

	k := ecdsa.PublicKey{Curve, x, y}
	enc := EncodePublicKey(k)
	if enc != expected {
		t.Errorf("expected encoding %q got %q", expected, enc)
	}

	kdec, err := DecodePublicKey(enc)
	if err != nil {
		t.Errorf("error decoding: %s", err)
	} else if !reflect.DeepEqual(*kdec, k) {
		t.Errorf("expected decoded key %v, got %v", k, *kdec)
	}
}

func TestEncodeDecodePrivateKey(t *testing.T) {
	dstr := "110061533099052216443366667375830438989694075383357356202709338247091019586025948858339719782647856308793524137003493025864000964698410222186380002561296076"
	expected := "CDVxkUebjoVK4WLI48RODZjmswH-cfRitXVQbpi2ne_xHG8JSuifIoD654zjpWClbMegyvPKXr8En1k2ZETfBsw="
	d, _ := big.NewInt(0).SetString(dstr, 10)

	k := ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: Curve}, D: d}
	enc := EncodePrivateKey(k)
	if enc != expected {
		t.Errorf("expected %q got %q", expected, enc)
	}

	kdec, err := DecodePrivateKey(enc)
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
	key64 := "CDVxkUebjoVK4WLI48RODZjmswH-cfRitXVQbpi2ne_xHG8JSuifIoD654zjpWClbMegyvPKXr8En1k2ZETfBsw="
	kdec, err := DecodePrivateKey(key64)
	if err != nil {
		t.Errorf("error decoding: %s", err)
	}

	_, err = SignUser("omgwtfbbq", kdec)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestVerifySignedUser(t *testing.T) {
	for i := 0; i < 20; i++ {
		priv := genTestKey(t)
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
			t.Fatalf("verification failed (%d)", i)
		}
	}
}

func genTestKey(t *testing.T) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(Curve, rand.Reader)
	if err != nil {
		t.Fatalf("error generating test key: %s", err)
	}
	return k
}
