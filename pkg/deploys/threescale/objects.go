package threescale

import (
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
)

// 3scale plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "3scale",
			ID:          "3scale-service-id",
			Description: "3scale",
			Metadata:    map[string]string{"serviceName": "3scale", "serviceType": "3scale"},
			Plans: []brokerapi.ServicePlan{
				{
					Name:        "default-3scale",
					ID:          "default-3scale",
					Description: "default 3scale plan",
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
