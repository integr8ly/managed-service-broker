package test_suites

import (
	"fmt"
	brokerClient "github.com/aerogear/managed-services-broker/tests/broker_client"
	"net/http"
	"sync"
	"testing"
	"time"
)

const (
	DEFAULT_OPERATION = "last"
	TEST_NAMESPACE = "test1"
)

func BrokerOperationsSuite(t *testing.T, svcs []*brokerClient.Service, sbc *brokerClient.ServiceBrokerClient, async bool) {

	for _, svc := range svcs {
		// Test CreateServiceInstance
		res, csiRes, err := sbc.CreateServiceInstance(TEST_NAMESPACE, svc.Name, svc.ID, svc.Plans[0].ID, async);if err != nil {
			t.Fatal(fmt.Sprintf("Unexpected error creating %s", svc.Name), err.Description)
		}
		if res.StatusCode != http.StatusAccepted {
			t.Fatal(fmt.Sprintf("Unexpected http status code when creating %s asynchronously", svc.Name), res.StatusCode, err.Description)
		}

		// Service Brokers MAY return an identifier representing the operation.
		// If not returned set to default as it can not be an empty string.
		operation := csiRes.Operation
		if operation == "" {
			operation = DEFAULT_OPERATION
		}

		// Poll for successful completion of CreateServiceInstance using LastOperation
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func(){
			defer wg.Done()
			hasFinished := false
			for hasFinished == false {
				_, loRes, _ := sbc.GetServiceInstanceLastOperation(svc.Name, svc.ID, svc.Plans[0].ID, operation);
				time.Sleep(time.Second * 1)
				// Ignore errors/error states as GetServiceInstanceLastOperation can return errors/error states as operations are initialising.
				// Continue polling until the Service Broker returns a valid response or the test timeout fails.
				hasFinished = loRes != nil && loRes.State == brokerClient.StateSucceeded
			}
		}()
		wg.Wait()

		// Test RemoveServiceInstance
		_, _, err = sbc.DeleteServiceInstance(svc.Name, svc.ID, svc.Plans[0].ID, async);if err != nil {
			t.Fatal(fmt.Sprintf("An error has occured deleting a service instance for %s", svc.Name), err.Description)
		}
	}
}