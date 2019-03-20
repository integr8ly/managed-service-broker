package fuse_managed

import (
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
)

// fuse managed plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "fuse-managed",
			ID:          "fuse-managed-service-id",
			Description: "fuse managed",
			Metadata:    map[string]string{"serviceName": "fuse-managed", "serviceType": "fuse"},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-fuse-managed",
					ID:          "default-fuse-managed",
					Description: "default fuse-managed plan",
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
