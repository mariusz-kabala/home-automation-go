package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const api = "https://api.godaddy.com/v1"

var apiKey string
var keySecret string

type domainDefinition struct {
	domain   string
	hostname string
}

type dnsRecord struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

type updatePayload struct {
	Data string `json:"data"`
}

var domains = []domainDefinition{
	domainDefinition{"kabala.tech", "@"},
	domainDefinition{"kabala.tech", "jenkins"},
	domainDefinition{"kabala.tech", "jklawyers"},
	domainDefinition{"geotags.pl", "@"},
	domainDefinition{"geotags.pl", "*"},
	domainDefinition{"geotags.eu", "@"},
	domainDefinition{"kabala.biz", "@"},
	domainDefinition{"kabala.eu", "@"},
}

func getApiKey() string {
	if apiKey != "" {
		return apiKey
	}

	return os.Getenv("API_KEY")
}

func getKeySecret() string {
	if keySecret != "" {
		return keySecret
	}

	return os.Getenv("KEY_SECRET")
}

func updateDomain(ddefinition domainDefinition, ip string) bool {
	payload := [1]updatePayload{
		updatePayload{ip},
	}

	json, err := json.Marshal(payload)

	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/domains/%s/records/A/%s", api, ddefinition.domain, ddefinition.hostname), bytes.NewBuffer(json))

	if err != nil {
		panic("Can not make http request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", getApiKey(), getKeySecret()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		panic("Error during making a http request")
	}

	return resp.StatusCode == 200
}

func fetchDomainInfo(ddefinition domainDefinition) []dnsRecord {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/domains/%s/records/A/%s", api, ddefinition.domain, ddefinition.hostname), nil)

	if err != nil {
		panic("Can not make http request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", getApiKey(), getKeySecret()))

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		panic("Error during making a http request")
	}

	var dns []dnsRecord

	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	jsonErr := json.Unmarshal(content, &dns)

	if jsonErr != nil {
		panic("Can not parse response json")
	}

	for _, record := range dns {
		fmt.Println(record.Data)
	}

	return dns
}

func getExternalIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)

	return string(content)
}

func init() {
	flag.StringVar(&apiKey, "apiKey", "", "GoDaddy API Key")
	flag.StringVar(&keySecret, "keySecret", "", "GoDaddy API Key Secret")

	flag.Parse()
}

func main() {
	godotenv.Load()

	currentIP := getExternalIP()

	fmt.Printf("Your IP Address is: %s\n", currentIP)

	for _, domain := range domains {
		dns := fetchDomainInfo(domain)

		if dns[0].Data != currentIP {
			if updateDomain(domain, currentIP) {
				fmt.Printf("Domain %s, record %s has been updated with ip address %s\n", domain.domain, domain.hostname, currentIP)
			} else {
				fmt.Printf("Can not update domain %s, record %s. Invalid response status code", domain.domain, domain.hostname)
			}
		} else {
			fmt.Printf("Domain %s, record %s is up to date\n", domain.domain, domain.hostname)
		}

	}
}
