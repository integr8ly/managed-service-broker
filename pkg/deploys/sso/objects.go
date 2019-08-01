package sso

import (
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
)

func getManagedCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "rhsso",
			ID:          "rhsso-service-id",
			Description: "RH-SSO",
			Metadata:    map[string]string{"serviceName": "rhsso", "serviceType": "rhsso"},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-rhsso",
					ID:          "default-rhsso",
					Description: "default rhsso plan",
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

func getUserCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "user-rhsso",
			ID:          "user-rhsso-service-id",
			Description: "User RH-SSO",
			Metadata:    map[string]string{"serviceName": "user-rhsso", "serviceType": "rhsso"},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-user-rhsso",
					ID:          "default-user-rhsso",
					Description: "default user rhsso plan",
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
