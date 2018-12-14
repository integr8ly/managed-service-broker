package fuse

import (
	testapi "github.com/integr8ly/managed-service-broker/tests/apis"
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
	"testing"
)

const (
	FUSE_NAME               = "fuse"
	MANAGED_SERVICE_PROJECT = "managed-service-project"
)

func SharedServiceSuite(t *testing.T, tc *testapi.TestCase, cCfg *brokerClient.BrokerClientClientConfig) {
	if tc.Service.Name != FUSE_NAME {
		t.Fatal("This test suite is for fuse only")
	}

	sbc := brokerClient.NewServiceBrokerClient(cCfg)
	for _, p := range tc.Service.Plans {
		TestSharedFuse(t, tc, &p, MANAGED_SERVICE_PROJECT, sbc)
	}
}
