package che

import (
	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
)

// Che plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "che",
			ID:          "che-service-id",
			Description: "che",
			Metadata:    map[string]string{"serviceName": "che", "serviceType": "che"},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					Name:        "default-che",
					ID:          "default-che",
					Description: "default che plan",
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
