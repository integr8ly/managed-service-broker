package launcher

import (
	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
)

// Launcher plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "launcher",
			ID:          "launcher-service-id",
			Description: "launcher",
			Metadata:    map[string]string{"serviceName": "launcher", "serviceType": "launcher"},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					Name:        "default-launcher",
					ID:          "default-launcher",
					Description: "default launcher plan",
					Free:        true,
					Schemas: &brokerapi.Schemas{
						ServiceBinding: &brokerapi.ServiceBindingSchema{
							Create: &brokerapi.RequestResponseSchema{},
						},
						ServiceInstance: &brokerapi.ServiceInstanceSchema{},
					},
				},
			},
		},
	}
}
