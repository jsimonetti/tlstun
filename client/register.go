package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
	"github.com/jsimonetti/tlstun/server"
	"github.com/spf13/viper"
)

func (c *client) registerPost(pass string) (string, error) {
	uri := fmt.Sprintf("https://%s/tlstun/register", viper.GetString("client_serveraddress"))

	tr := &http.Transport{
		TLSClientConfig: c.tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}
	hc := http.Client{Transport: tr}

	form := url.Values{}
	form.Add("password", pass)

	resp, err := hc.PostForm(uri, form)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", s), err
}

func (c *client) RegisterStatus() {
	uri := fmt.Sprintf("https://%s/tlstun/status", viper.GetString("client_serveraddress"))

	tr := &http.Transport{
		TLSClientConfig: c.tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}
	hc := http.Client{Transport: tr}

	resp, err := hc.Get(uri)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}
	defer resp.Body.Close()

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error: %s", err)
		return
	}

	response := fmt.Sprintf("%s", s)
	if response == server.TrustedResponse() {
		fmt.Print("You are registered\n")
		return
	}
	if response == server.UnTrustedResponse() {
		fmt.Print("You are not registered\n")
		return
	}

	fmt.Printf("Unknown error: %s\n", response)
}

func (c *client) Register() {
	var password string
	fmt.Print("Enter password:")
	pwd, _ := gopass.GetPasswd()

	password = string(pwd)
	resp, err := c.registerPost(password)

	if err != nil {
		fmt.Print("\nRegistration failed", err)
		return
	}
	fmt.Printf("\nResponse: %s\n", resp)
}
