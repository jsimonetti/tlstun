package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"

	"github.com/jsimonetti/tlstun/shared"
)

var scert *x509.Certificate

func register() error {
	var password string
	fmt.Printf("Enter password:")
	pwd := gopass.GetPasswd()

	password = string(pwd)
	resp, _ := post(password)

	fmt.Printf("\nResponse: %s\n", resp)

	return nil
}

func post(pass string) (string, error) {
	mynil := ""
	certf, keyf, err := readMyCert()
	if err != nil {
		return mynil, err
	}
	tlsConfig, err := shared.GetTLSConfig(certf, keyf)
	if err != nil {
		return mynil, err
	}

	//loadServerCert()
	uri := fmt.Sprintf("https://%s:%d/register", serverIp, serverPort)

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}
	hc := http.Client{Transport: tr}

	form := url.Values{}
	form.Add("password", pass)

	resp, err := hc.PostForm(uri, form)
	if err != nil {
		return mynil, err
	}
	defer resp.Body.Close()
	s, err := ioutil.ReadAll(resp.Body)
	val := fmt.Sprintf("%s", s)
	if err != nil {
		return mynil, err
	}
	return val, err
}

func loadServerCert() {
	name := fmt.Sprintf("server-%s:%d.crt", serverIp, serverPort)
	cert, err := ReadCert(name)
	if err != nil {
		shared.Log("client", "error", fmt.Sprintf("Error reading the server certificate for %s: %v", name, err))
		return
	}

	scert = cert
}

func ReadCert(fpath string) (*x509.Certificate, error) {
	cf, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(cf)
	return x509.ParseCertificate(certBlock.Bytes)
}
