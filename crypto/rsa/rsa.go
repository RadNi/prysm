package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	enginev1 "github.com/prysmaticlabs/prysm/v3/proto/engine/v1"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/log"
)

func ExportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) string {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return ""
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem)
}

func ExportRsaPrivateKeyAsPemStr(privkey *rsa.PrivateKey) string {
	privkey_bytes := x509.MarshalPKCS1PrivateKey(privkey)
	privkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privkey_bytes,
		},
	)
	return string(privkey_pem)
}

func ParseRsaPrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

func ParseRsaPublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("Key type is not RSA")
}

func ExportPrivatekey(prv *rsa.PrivateKey) {
	fi, err := os.Create("prv.pem")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	prvPEM := ExportRsaPrivateKeyAsPemStr(prv)
	num, _ := fi.WriteString(prvPEM)
	fmt.Printf("Wrote %d bytes\n", num)
}

func ExportPublickey(pub *rsa.PublicKey) {
	fi, err := os.Create("pub.pem")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	pubPEM := ExportRsaPublicKeyAsPemStr(pub)
	num, _ := fi.WriteString(pubPEM)
	fmt.Printf("Wrote %d bytes\n", num)
}

func ImportPublicKey() *rsa.PublicKey {
	b, err := os.ReadFile("pub.pem") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	pubPEM := string(b)
	log.Info("rsa:")
	log.Info(pubPEM)
	pub, _ := ParseRsaPublicKeyFromPemStr(pubPEM)
	return pub
}

func ImportPrivateKey() *rsa.PrivateKey {
	b, err := os.ReadFile("prv.pem") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	prvPEM := string(b)
	prv, _ := ParseRsaPrivateKeyFromPemStr(prvPEM)
	return prv
}

func ToProtoRSAPrivatekey(prv *rsa.PrivateKey) *enginev1.RSAPrivateKey {
	return NewProtoPrivateKey(prv.D, prv.Primes, prv.PublicKey.N, prv.PublicKey.E)
}

func ToProtoRSAPublickey(pub *rsa.PublicKey) *enginev1.RSAPublicKey {
	return NewProtoPublicKey(pub.N, pub.E)
}

func NewProtoPublicKey(n *big.Int, e int) *enginev1.RSAPublicKey {
	return &enginev1.RSAPublicKey{
		N: n.Bytes(),
		E: uint64(e),
	}
}

func NewProtoPrivateKey(d *big.Int, p []*big.Int, n *big.Int, e int) *enginev1.RSAPrivateKey {
	primes := make([][]byte, len(p))
	for i, prime := range p {
		primes[i] = prime.Bytes()
	}
	return &enginev1.RSAPrivateKey{
		PublicKey: NewProtoPublicKey(n, e),
		D:         d.Bytes(),
		Primes:    primes,
	}
}

func FromProtoRSAPrivatekey(prv *enginev1.RSAPrivateKey) *rsa.PrivateKey {
	return NewPrivateKey(prv.D, prv.Primes, prv.PublicKey.N, prv.PublicKey.E)
}

func FromProtoRSAPublickey(pub *enginev1.RSAPublicKey) *rsa.PublicKey {
	return NewPublicKey(pub.N, pub.E)
}

func NewPublicKey(n []byte, e uint64) *rsa.PublicKey {
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(n),
		E: int(e),
	}
}

func NewPrivateKey(d []byte, p [][]byte, n []byte, e uint64) *rsa.PrivateKey {
	D := new(big.Int).SetBytes(d)
	primes := make([]*big.Int, len(p))
	for i, prime := range p {
		primes[i] = new(big.Int).SetBytes(prime)
	}
	return &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(e),
		},
		D:      D,
		Primes: primes,
		Precomputed: rsa.PrecomputedValues{
			Dp:        nil,
			Dq:        nil,
			Qinv:      nil,
			CRTValues: nil,
		},
	}
}

func KeyGen() (*rsa.PrivateKey, error) {
	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader
	prv, err := rsa.GenerateKey(rng, 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from keygen: %s\n", err)
		return nil, err
	}
	fmt.Printf("PrivateKey: %v\nPublicKey: %v\nPrimes: %v\n Precomputed.Dp: %v\n Precomputed.Dq: %v\n Precomputed.Qinv: %v\n", prv.D, prv.PublicKey, prv.Primes, &prv.Precomputed.Dp, &prv.Precomputed.Dq, &prv.Precomputed.Qinv)
	return prv, nil
}

func Encrypt(message []byte, pk *rsa.PublicKey) ([]byte, error) {
	label := []byte("orders")

	// crypto/rand.Reader is a good source of entropy for randomizing the
	// encryption function.
	rng := rand.Reader
	fmt.Printf("%d\n", pk.Size()-2*sha256.New().Size()-2)
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, pk, message, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from encryption: %s\n", err)
		return nil, err
	}

	// Since encryption is a randomized function, ciphertext will be
	// different each time.
	fmt.Printf("Ciphertext: %x\n", ciphertext)
	return ciphertext, nil
}

func Decrypt(ciphertext []byte, skt *enginev1.RSAPrivateKey) ([]byte, error) {
	sk := FromProtoRSAPrivatekey(skt)
	rng := rand.Reader
	label := []byte("orders")
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, sk, ciphertext, label)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from decryption: %s\n", err)
		return nil, err
	}
	return plaintext, nil
}

func EncryptMulti(msg []byte, pk *rsa.PublicKey) ([]byte, error) {
	msgLen := len(msg)
	step := pk.Size() - 2*sha256.New().Size() - 2
	var encryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encryptedBlockBytes, err := Encrypt(msg[start:finish], pk)
		if err != nil {
			return nil, err
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)
	}

	return encryptedBytes, nil
}

func DecryptMulti(ciphertext []byte, sk *enginev1.RSAPrivateKey) ([]byte, error) {
	msgLen := len(ciphertext)
	step := FromProtoRSAPublickey(sk.PublicKey).Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := Decrypt(ciphertext[start:finish], sk)
		if err != nil {
			return nil, err
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}

//func main() {
//	message := []byte("send reinforcements, we're going to advance")
//	prv := KeyGen()
//	publicPEM, _ := ExportRsaPublicKeyAsPemStr(&prv.PublicKey)
//	privatePEM := ExportRsaPrivateKeyAsPemStr(prv)
//	ExportPrivatekey(prv)
//	fmt.Printf("public pem: %s\nprivate pem: %s\n", publicPEM, privatePEM)
//	prv2 := ImportPrivateKey()
//	ciphertext := Encrypt(message, &prv.PublicKey)
//	plaintext := Decrypt(ciphertext, prv)
//	fmt.Printf("%d %v\n", len(plaintext), string(plaintext))
//	plaintext2 := Decrypt(ciphertext, prv2)
//	fmt.Printf("%d %v\n", len(plaintext2), string(plaintext2))
//}
