package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

var selfSignedTLSCertificate []byte = []byte{}
var selfSignedTLSKey []byte = []byte{}

func GetTLSCertificate() []byte {
	return selfSignedTLSCertificate
}

func GetTLSKey() []byte {
	return selfSignedTLSKey
}

func generateSelfSignedTLSCertificate() error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "ablage",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("Failed to create new x509 certificate: %v", err)
	}

	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	key, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("Failed to marshal EC private key: %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: key})

	selfSignedTLSCertificate = cert
	selfSignedTLSKey = keyPEM

	return nil
}

func loadOrGenerateTLSCertificate() error {
	if GetHttpMode() {
		return nil
	}

	if pathTLSCertFile == "" || pathTLSKeyFile == "" {
		return generateSelfSignedTLSCertificate()
	}

	_, err := tls.LoadX509KeyPair(pathTLSCertFile, pathTLSKeyFile)
	if err != nil {
		return fmt.Errorf("Failed to load TLS certificate or key: %w", err)
	}

	certData, err := os.ReadFile(pathTLSCertFile)
	if err != nil {
		return fmt.Errorf("Failed to read TLS certificate file: %w", err)
	}

	keyData, err := os.ReadFile(pathTLSKeyFile)
	if err != nil {
		return fmt.Errorf("Failed to read TLS key file: %w", err)
	}

	selfSignedTLSCertificate = certData
	selfSignedTLSKey = keyData

	return nil
}
