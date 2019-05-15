package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"log"
)

var baseURL = os.Getenv("BASE_URL")

const (
	// analytics layer
	endpointPostClassificationTwitter = "/ri-analytics-classification-twitter/lang/"

	// collection layer
	endpointGetCrawlTweets              = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s/fast"
	endpointGetCrawlAllAvailableTweets  = "/ri-collection-explicit-feedback-twitter/mention/%s/lang/%s"
	endpointGetTwitterAccountNameExists = "/ri-collection-explicit-feedback-twitter/%s/exists"

	// storage layer
	endpointPostObserveTwitterAccount     = "/ri-storage-twitter/store/observable/"
	endpointGetObservablesTwitterAccounts = "/ri-storage-twitter/observables"
	endpointGetUnclassifiedTweets         = "/ri-storage-twitter/account_name/%s/lang/%s/unclassified"
	endpointPostTweet                     = "/ri-storage-twitter/store/tweet/"
	endpointPostClassifiedTweet           = "/ri-storage-twitter/store/classified/tweet/"
)

var client = getHTTPClient()

func getHTTPClient() *http.Client {
	pwd, _ := os.Getwd()
	caCert, err := ioutil.ReadFile(pwd + "/ca_chain.crt")
	timeout := time.Duration(2 * time.Minute)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
		Timeout: timeout,
	}

	return client
}

// RESTPostStoreObserveTwitterAccount returns ok
func RESTPostStoreObserveTwitterAccount(obserable ObservableTwitter) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(obserable)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostObserveTwitterAccount
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR post store observable %v\n", err)
	}
	if res.StatusCode == 200 {
		return true
	}

	return false
}

// RESTGetObservablesTwitterAccounts retrieve all observables from the storage layer
func RESTGetObservablesTwitterAccounts() []ObservableTwitter {
	var obserables []ObservableTwitter

	url := baseURL + endpointGetObservablesTwitterAccounts
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send observable account get request", err)
		return obserables
	}

	err = json.NewDecoder(res.Body).Decode(&obserables)
	if err != nil {
		fmt.Println("ERR cannot decode twitter observable json", err)
		return obserables
	}

	return obserables
}

// RESTGetTwitterAccountNameExists check if a twitter account exists
func RESTGetTwitterAccountNameExists(accountName string) CrawlerResponseMessage {
	var response CrawlerResponseMessage

	endpoint := fmt.Sprintf(endpointGetTwitterAccountNameExists, accountName)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send get request to check if Twitter account exists", err)
		return response
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		fmt.Println("ERR cannot decode response from the twitter crawler json", err)
		return response
	}

	return response
}

// RESTGetUnclassifiedTweets retrieve all tweets from a specified account that have not been classified yet
func RESTGetUnclassifiedTweets(accountName, lang string) []Tweet {
	var tweet []Tweet

	endpoint := fmt.Sprintf(endpointGetUnclassifiedTweets, accountName, lang)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send get request to get unclassified tweets", err)
		return tweet
	}

	err = json.NewDecoder(res.Body).Decode(&tweet)
	if err != nil {
		fmt.Println("ERR cannot decode unclassified tweets json", err)
		return tweet
	}

	return tweet
}

// RESTGetCrawlTweets retrieve all tweets from the collection layer that addresses the given account name
func RESTGetCrawlTweets(accountName string, lang string) []Tweet {
	var tweets []Tweet

	endpoint := fmt.Sprintf(endpointGetCrawlTweets, accountName, lang)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR cannot send request to tweet crawler", err)
		return tweets
	}

	err = json.NewDecoder(res.Body).Decode(&tweets)
	if err != nil {
		fmt.Println("ERR cannot decode crawled tweets", err)
		return tweets
	}

	return tweets
}

func RESTGetCrawlMaximumNumberOfTweets(accountName string, lang string) []Tweet {
	var tweets []Tweet

	endpoint := fmt.Sprintf(endpointGetCrawlAllAvailableTweets, accountName, lang)
	url := baseURL + endpoint
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERR crawl max number of tweets", err)
		return tweets
	}

	err = json.NewDecoder(res.Body).Decode(&tweets)
	if err != nil {
		fmt.Println("ERR cannot decode tweets", err)
		return tweets
	}

	return tweets
}

// RESTPostClassifyTweets returns ok
func RESTPostClassifyTweets(tweets []Tweet, lang string) []Tweet {
	var classifiedTweets []Tweet

	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostClassificationTwitter + lang
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR %v\n", err)
	}

	err = json.NewDecoder(res.Body).Decode(&classifiedTweets)
	if err != nil {
		log.Printf("ERR cannot decode classified tweets %v\n", err)
	}

	return classifiedTweets
}

// RESTPostStoreTweets returns ok
func RESTPostStoreTweets(tweets []Tweet) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostTweet
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR cannot send request to store tweets %v\n", err)
	}
	if res.StatusCode == 200 {
		return true
	}

	return false
}

// RESTPostStoreClassifiedTweets returns ok
func RESTPostStoreClassifiedTweets(tweets []Tweet) bool {
	requestBody := new(bytes.Buffer)
	err := json.NewEncoder(requestBody).Encode(tweets)
	if err != nil {
		log.Printf("ERR - json formatting error: %v\n", err)
	}

	url := baseURL + endpointPostClassifiedTweet
	res, err := client.Post(url, "application/json; charset=utf-8", requestBody)
	if err != nil {
		log.Printf("ERR cannot send request to store tweets %v\n", err)
	}
	if res.StatusCode == 200 {
		return true
	}

	return false
}
