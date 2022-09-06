package main

import (
	"net/http"
	"log"
	"encoding/json"
	"fmt"
	"errors"
	"bytes"
	"time"
	"flag"
	"os"
)

var IP_PROVIDER = "https://v4.ident.me/"

func getOwnIPv4() (string, error) {
	resp, err := http.Get(IP_PROVIDER)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return buf.String(), nil
}

func getDomainIPv4() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A/%s", DOMAIN, SUBDOMAIN), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}
	in := make([]struct {
		Data string `json:"data"`
	}, 1)
	json.NewDecoder(resp.Body).Decode(&in)
	return in[0].Data, nil
}

func putNewIP(ip string) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(&[1]struct {
		Name string `json:"name"`
		Data string `json:"data"`
		TTL int64 `json:"ttl"`
	} {
		{
			SUBDOMAIN,
			ip,
			600,
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT",
		fmt.Sprintf("https://api.godaddy.com/v1/domains/%s/records/A", DOMAIN),
		&buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", GODADDY_KEY, GODADDY_SECRET))
	c := new(http.Client)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode == 200 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Failed with HTTP status code %d\n", resp.StatusCode))
	}
}

func run() {
	ownIP, err := getOwnIPv4()
	if err != nil {
		log.Fatal(err)
	}
	domainIP, err := getDomainIPv4()
	if err != nil {
		log.Fatal(err)
	}
	if domainIP != ownIP {
		if err := putNewIP(ownIP); err != nil {
			log.Fatal(err)
		}
	}
}

// globals
var GODADDY_KEY = ""
var GODADDY_SECRET = ""
var DOMAIN = ""
var SUBDOMAIN = ""

func main() {
	// log file flag
	logFile := flag.String("log", "", "Path for log file (will be created if it doesn't exist)")
	// required flags
	keyPtr := flag.String("key", "", "Godaddy API key")
	secretPtr := flag.String("secret", "", "Godaddy API secret")
	domainPtr := flag.String("domain", "", "Your top level domain (e.g., example.com) registered with Godaddy and on the same account as your API key")
	// optional flags
	subdomainPtr := flag.String("subdomain", "@", "The data value (aka host) for the A record. It can be a 'subdomain' (e.g., 'subdomain' where 'subdomain.example.com' is the qualified domain name). Note that such an A record must be set up first in your Godaddy account beforehand. Defaults to @. (Optional)")
	POLLING := flag.Int64("interval", 360, "Polling interval in seconds. Lookup Godaddy's current rate limits before setting too low. Defaults to 360. (Optional)")

	flag.Parse()
	SUBDOMAIN = *subdomainPtr
	DOMAIN = *domainPtr
	GODADDY_SECRET = *secretPtr
	GODADDY_KEY = *keyPtr

	if *logFile == "" {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile(*logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Couldn't open log file: %s", err)
		}
		defer f.Close()
		log.SetOutput(f)
	}

	if DOMAIN == "" {
		log.Fatalf("You need to provide your domain")
	}

	if GODADDY_SECRET == "" {
		log.Fatalf("You need to provide your API secret")
	}

	if GODADDY_KEY == "" {
		log.Fatalf("You need to provide your API key")
	}

	// run
	for {
		run()
		time.Sleep(time.Second * time.Duration(*POLLING))
	}
}
