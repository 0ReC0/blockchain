package p2p

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

func LoadTLSCert(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}

func GenerateTLSConfig() *tls.Config {
	serverCert, err := LoadTLSCert("certs/server.crt", "certs/server.key")
	if err != nil {
		panic(err)
	}

	// Загрузка CA для проверки клиентов
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}
}
