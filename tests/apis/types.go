package apis

import (
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
)

type TestCase struct {
	Service   *brokerClient.Service
	Namespace string
	Async     bool
}
