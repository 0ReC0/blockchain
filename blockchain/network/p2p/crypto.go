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

func GenerateTLSConfig(nodeAddress string) *tls.Config {
	var certFile, keyFile string
	switch nodeAddress {
	case "localhost:26656":
		certFile = "certs/validator1.crt"
		keyFile = "certs/validator1.key"
	case "localhost:26657":
		certFile = "certs/validator2.crt"
		keyFile = "certs/validator2.key"
	case "localhost:26658":
		certFile = "certs/validator3.crt"
		keyFile = "certs/validator3.key"
	case "localhost:26659":
		certFile = "certs/validator4.crt"
		keyFile = "certs/validator4.key"
	case "localhost:26660":
		certFile = "certs/validator5.crt"
		keyFile = "certs/validator5.key"
	default:
		panic("no certificate for address: " + nodeAddress)
	}
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
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
		MinVersion:   tls.VersionTLS12,
		ServerName:   "mydomain.local",
	}
}

func GenerateClientTLSConfig() *tls.Config {
	// Загрузка CA
	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		panic(err)
	}
	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	// Получаем NODE_ADDRESS
	nodeAddr := os.Getenv("NODE_ADDRESS")
	if nodeAddr == "" {
		panic("NODE_ADDRESS is not set")
	}

	// Загрузка клиентского сертификата
	var clientCert tls.Certificate
	switch nodeAddr {
	case "localhost:26656":
		clientCert, err = tls.LoadX509KeyPair("certs/validator1.crt", "certs/validator1.key")
	case "localhost:26657":
		clientCert, err = tls.LoadX509KeyPair("certs/validator2.crt", "certs/validator2.key")
	case "localhost:26658":
		clientCert, err = tls.LoadX509KeyPair("certs/validator3.crt", "certs/validator3.key")
	case "localhost:26659":
		clientCert, err = tls.LoadX509KeyPair("certs/validator4.crt", "certs/validator4.key")
	case "localhost:26660":
		clientCert, err = tls.LoadX509KeyPair("certs/validator5.crt", "certs/validator5.key")
	default:
		panic("unknown node address: " + nodeAddr)
	}
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   "mydomain.local", // совпадает с SAN
	}
}
