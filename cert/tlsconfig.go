package cert

import "crypto/tls"

func TLSConfig(certf string, keyf string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certf, keyf)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequestClientCert,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		MaxVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		PreferServerCipherSuites: true,
	}

	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
