package proxy

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"sync"
	"time"
)

var certMap = &sync.Map{}

func createCert(dnsName string, parent *x509.Certificate, parentKey crypto.PrivateKey, hoursValid int) (cert []byte, priv []byte, err error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to generate private key: %v", err))
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to generate serial number: %v", err))
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"MITM proxy"},
		},
		DNSNames:  []string{dnsName},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Duration(hoursValid) * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, parent, &privateKey.PublicKey, parentKey)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Failed to create certificate: %v", err))
	}
	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		return nil, nil, errors.New(fmt.Sprintf("failed to encode certificate to PEM"))
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, errors.New(fmt.Sprintf("Unable to marshal private key: %v", err))
	}
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if len(pemKey) == 0 {
		return nil, nil, errors.New(fmt.Sprintf("failed to encode key to PEM"))
	}

	return pemCert, pemKey, nil
}

func loadX509KeyPair(certFile, keyFile string) (cert *x509.Certificate, key any, err error) {
	cf, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, err
	}

	kf, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, err
	}
	certBlock, _ := pem.Decode(cf)
	cert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	keyBlock, _ := pem.Decode(kf)
	key, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func getTlsCert(host string, parent *x509.Certificate, parentKey crypto.PrivateKey) *tls.Certificate {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		slog.Error("error splitting h/port", slog.String("h", h), slog.Any("err", err))
		h = host
	}

	v, ok := certMap.Load(h)
	var tlsCert tls.Certificate
	if ok {
		tlsCert = v.(tls.Certificate)
	} else {
		pemCert, pemKey, err := createCert(h, parent, parentKey, 240)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to create cert/key pair: %v", err))
			return nil
		}
		tlsCert, err = tls.X509KeyPair(pemCert, pemKey)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to create certificate: %v", err))
			return nil
		}

		certMap.Store(h, tlsCert)
	}

	return &tlsCert
}
