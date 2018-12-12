package broker

import (
	"fmt"
	testapi "github.com/integr8ly/managed-service-broker/tests/apis"
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
	"net/http"
	"testing"
)

const (
	DEFAULT_OPERATION = "operation"
)

func TestProvision(t *testing.T, tc *testapi.TestCase, svcp *brokerClient.ServicePlan, sbc *brokerClient.ServiceBrokerClient) {

	res, csiRes, err := sbc.CreateServiceInstance(tc.Namespace, tc.Service.Name, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error creating %s", tc.Service.Name), err.Description)
	}

	if tc.Async == false && res.StatusCode != http.StatusCreated {
		t.Fatal(fmt.Sprintf("Unexpected http status code when creating %s synchronously", tc.Service.Name), res.StatusCode, err.Description)
	}

	if tc.Async == true {
		if res.StatusCode != http.StatusAccepted {
			t.Fatal(fmt.Sprintf("Unexpected http status code when creating %s asynchronously", tc.Service.Name), res.StatusCode, err.Description)
		}

		err := confirmLastOperation(tc.Service.Name, tc.Service, svcp, csiRes.Operation, sbc)
		if err != nil {
			t.Fatal("Error polling last operation", err)
		}
	}
}

func TestDeprovision(t *testing.T, tc *testapi.TestCase, svcp *brokerClient.ServicePlan, sbc *brokerClient.ServiceBrokerClient) {

	res, dsiRes, err := sbc.DeleteServiceInstance(tc.Service.Name, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("An error has occured deleting a service instance for %s", tc.Service.Name), err.Description)
	}

	if tc.Async == false && res.StatusCode != http.StatusOK {
		t.Fatal(fmt.Sprintf("Unexpected http status code when deleting %s synchronously", tc.Service.Name), res.StatusCode, err.Description)
	}

	if tc.Async == true {
		if res.StatusCode != http.StatusAccepted {
			t.Fatal(fmt.Sprintf("Unexpected http status code when deleting %s asynchronously", tc.Service.Name), res.StatusCode, err.Description)
		}

		err := confirmLastOperation(tc.Service.Name, tc.Service, svcp, dsiRes.Operation, sbc)
		if err != nil {
			t.Fatal("Error polling last operation", err)
		}
	}
}

func confirmLastOperation(instanceId string, svc *brokerClient.Service, svcp *brokerClient.ServicePlan, operation string, sbc *brokerClient.ServiceBrokerClient) error {
	// Service Brokers MAY return an identifier representing the operation.
	// If not returned set to default as it can not be an empty string.
	if operation == "" {
		operation = DEFAULT_OPERATION
	}

	// PollLastOperation polls until a successful result is returned.
	// If there is no error the test timeout will fail the test.
	err := sbc.PollLastOperation(instanceId, svc, svcp, operation)
	if err != nil {
		return err
	}

	return nil
}
