package test

import (
	"flag"
	"fmt"
	brokerClient "github.com/aerogear/managed-services-broker/tests/broker_client"
	suites "github.com/aerogear/managed-services-broker/tests/test_suites"
	"os"
	"testing"
)

// https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md#originating-identity
// The Base64 encoded string matches the below JSON object.
// {
//  "username": "developer",
//  "uid": "",
//  "groups": ["system:authenticated:oauth", "system:authenticated"],
//  "extra": {
//    "scopes.authorization.openshift.io": ["user:full"]
//  }
//}
// The required format is `platform Base64value`.
const USER_IDENTITY = "kubernetes eyJ1c2VybmFtZSI6ImRldmVsb3BlciIsInVpZCI6IiIsImdyb3VwcyI6WyJzeXN0ZW06YXV0aGVudGljYXRlZDpvYXV0aCIsInN5c3RlbTphdXRoZW50aWNhdGVkIl0sImV4dHJhIjp7InNjb3Blcy5hdXRob3JpemF0aW9uLm9wZW5zaGlmdC5pbyI6WyJ1c2VyOmZ1bGwiXX19"

const (
	BROKER_URL               = "BROKER_URL"
	KUBERNETES_API_TOKEN     = "KUBERNETES_API_TOKEN"
)

var (
	envBrokerURL     = os.Getenv(BROKER_URL)
	envToken         = os.Getenv(KUBERNETES_API_TOKEN)
)

var (
	flagBrokerURL = flag.String(BROKER_URL, "", "URL of the manged-services-broker")
	flagToken     = flag.String(KUBERNETES_API_TOKEN, "", "Kubernetes token for authorisation")
)

func getBrokerDetails() (string, string) {
	brokerURL := *flagBrokerURL
	token := *flagToken
	if brokerURL == "" {
		brokerURL = os.Getenv(BROKER_URL)
	}
	if token == "" {
		token = os.Getenv(KUBERNETES_API_TOKEN)
	}

	return brokerURL, token
}

func TestBrokerOperations(t *testing.T) {
	brokerURL, token  := getBrokerDetails()
	if brokerURL == "" || token == "" {
		t.Fatal(fmt.Sprintf("Please make sure %s and %s are set before running tests.", BROKER_URL, KUBERNETES_API_TOKEN))
	}

	sbc := brokerClient.NewServiceBrokerClient(brokerURL, token, USER_IDENTITY, true)

	// Test catalog
	_, sc, err := sbc.Catalog();if err != nil {
		t.Fatal(fmt.Sprintf("Error getting Catalog: "), err.Description)
	}

	// Generic broker tests
	suites.BrokerOperationsSuite(t, sc.Services, sbc, true)
}