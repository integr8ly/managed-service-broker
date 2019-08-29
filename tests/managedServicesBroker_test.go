package test

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"testing"

	testapi "github.com/integr8ly/managed-service-broker/tests/apis"
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
	"github.com/integr8ly/managed-service-broker/tests/test_suites/broker"
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
	BROKER_URL           = "BROKER_URL"
	KUBERNETES_API_TOKEN = "KUBERNETES_API_TOKEN"
	// This must match a consumer namespace in deploy/cr.yaml
	TEST_NAMESPACE = "consumer1"
)

var (
	envBrokerURL = os.Getenv(BROKER_URL)
	envToken     = os.Getenv(KUBERNETES_API_TOKEN)
	numServices  = 9
)

var (
	flagBrokerURL = flag.String(BROKER_URL, "", "URL of the manged-service-broker")
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

func TestManagedBroker(t *testing.T) {
	brokerURL, token := getBrokerDetails()
	if brokerURL == "" || token == "" {
		t.Fatal(fmt.Sprintf("Please make sure %s and %s are set before running tests.", BROKER_URL, KUBERNETES_API_TOKEN))
	}

	cCfg := &brokerClient.BrokerClientClientConfig{
		brokerURL,
		token,
		USER_IDENTITY,
		&tls.Config{InsecureSkipVerify: true},
	}
	sbc := brokerClient.NewServiceBrokerClient(cCfg)

	// Test catalog
	res, sc, err := sbc.Catalog()
	if err != nil {
		message := err.Description
		if len(message) == 0 {
			message = res.Status
		}
		t.Fatal(fmt.Sprintf("Error getting Catalog: %s", message))
	}
	if len(sc.Services) != numServices {
		t.Fatalf("There should be %d managed services, but there are %d of them", numServices, len(sc.Services))
	}

	// Generic broker tests
	for _, svc := range sc.Services {
		broker.OperationsSuite(t, &testapi.TestCase{svc, TEST_NAMESPACE, true}, cCfg)
	}
}
