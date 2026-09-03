package ioscert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

var oidEmailAddress = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 1}

type CSRResult struct {
	KeyPath string
	CSRPath string
}

func GenerateKeyAndCSR(outDir, email, commonName, country string) (*CSRResult, error) {
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return nil, err
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	subject := pkix.Name{
		CommonName: commonName,
	}
	if country != "" {
		subject.Country = []string{country}
	}
	if email != "" {
		subject.ExtraNames = append(subject.ExtraNames, pkix.AttributeTypeAndValue{
			Type:  oidEmailAddress,
			Value: email,
		})
	}

	template := x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, &template, key)
	if err != nil {
		return nil, fmt.Errorf("creating CSR: %w", err)
	}

	keyPath := filepath.Join(outDir, "ios_distribution.key.pem")
	csrPath := filepath.Join(outDir, "CertificateSigningRequest.certSigningRequest")

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, fmt.Errorf("writing private key: %w", err)
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})
	if err := os.WriteFile(csrPath, csrPEM, 0644); err != nil {
		return nil, fmt.Errorf("writing CSR: %w", err)
	}

	return &CSRResult{KeyPath: keyPath, CSRPath: csrPath}, nil
}
