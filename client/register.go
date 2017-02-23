package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"

	"github.com/jsimonetti/tlstun/shared"
)

func register() error {
	var password string
	fmt.Print("Enter password:")
	pwd, _ := gopass.GetPasswd()

	password = string(pwd)
	resp, err := post(password)

	if err != nil {
		fmt.Print("\nRegistration failed", err)
		return err
	}
	fmt.Printf("\nResponse: %s\n", resp)

	return nil
}

func post(pass string) (string, error) {
	mynil := ""
	certf, keyf, err := shared.ReadMyCert("client.crt", "client.key")
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


