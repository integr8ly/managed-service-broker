package broker

import (
	testapi "github.com/integr8ly/managed-service-broker/tests/apis"
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
	"testing"
)

func OperationsSuite(t *testing.T, tc *testapi.TestCase, cCfg *brokerClient.BrokerClientClientConfig) {
	sbc := brokerClient.NewServiceBrokerClient(cCfg)
	for _, p := range tc.Service.Plans {
		TestProvision(t, tc, &p, sbc)
		TestDeprovision(t, tc, &p, sbc)
	}
}
