package sso

import (
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
)

// sso managed plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "rhsso",
			ID:          "rhsso-service-id",
			Description: "RH-SSO",
			Metadata:    map[string]string{"serviceName": "rhsso", "serviceType": "rhsso"},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-rhsso-managed",
					ID:          "default-rhsso-managed",
					Description: "default rhsso-managed plan",
					Free:        true,
					Schemas: &brokerapi.Schemas{
						ServiceBinding: &brokerapi.ServiceBindingSchema{
							Create: &brokerapi.RequestResponseSchema{},
						},
						ServiceInstance: &brokerapi.ServiceInstanceSchema{
							Create: &brokerapi.InputParametersSchema{},
						},
					},
				},
			},
		},
	}
}
