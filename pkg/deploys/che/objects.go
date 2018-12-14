package che

import (
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
)

// Che plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "che",
			ID:          "che-service-id",
			Description: "che",
			Metadata: map[string]string{
				"serviceName": "che",
				"serviceType": "che",
				"imageUrl":    "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+CjwhLS0KCiAgICBDb3B5cmlnaHQgKGMpIDIwMTUtMjAxOCBSZWQgSGF0LCBJbmMuCiAgICBUaGlzIHByb2dyYW0gYW5kIHRoZSBhY2NvbXBhbnlpbmcgbWF0ZXJpYWxzIGFyZSBtYWRlCiAgICBhdmFpbGFibGUgdW5kZXIgdGhlIHRlcm1zIG9mIHRoZSBFY2xpcHNlIFB1YmxpYyBMaWNlbnNlIDIuMAogICAgd2hpY2ggaXMgYXZhaWxhYmxlIGF0IGh0dHBzOi8vd3d3LmVjbGlwc2Uub3JnL2xlZ2FsL2VwbC0yLjAvCgogICAgU1BEWC1MaWNlbnNlLUlkZW50aWZpZXI6IEVQTC0yLjAKCiAgICBDb250cmlidXRvcnM6CiAgICAgIFJlZCBIYXQsIEluYy4gLSBpbml0aWFsIEFQSSBhbmQgaW1wbGVtZW50YXRpb24KCi0tPgo8c3ZnIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiBoZWlnaHQ9IjcwcHgiCiAgICAgd2lkdGg9IjcwcHgiIHZlcnNpb249IjEuMSIgeG1sbnM6Y2M9Imh0dHA6Ly9jcmVhdGl2ZWNvbW1vbnMub3JnL25zIyIgeG1sbnM6ZGM9Imh0dHA6Ly9wdXJsLm9yZy9kYy9lbGVtZW50cy8xLjEvIgogICAgIHZpZXdCb3g9IjAgMCA0NyA1NyI+CiAgPG1ldGFkYXRhPgogICAgPHJkZjpSREY+CiAgICAgIDxjYzpXb3JrIHJkZjphYm91dD0iIj4KICAgICAgICA8ZGM6Zm9ybWF0PmltYWdlL3N2Zyt4bWw8L2RjOmZvcm1hdD4KICAgICAgICA8ZGM6dHlwZSByZGY6cmVzb3VyY2U9Imh0dHA6Ly9wdXJsLm9yZy9kYy9kY21pdHlwZS9TdGlsbEltYWdlIi8+CiAgICAgICAgPGRjOnRpdGxlLz4KICAgICAgPC9jYzpXb3JrPgogICAgPC9yZGY6UkRGPgogIDwvbWV0YWRhdGE+CiAgPGcgZmlsbC1ydWxlPSJldmVub2RkIiBzdHJva2U9Im5vbmUiIHN0cm9rZS13aWR0aD0iMSIgZmlsbD0ibm9uZSI+CiAgICA8cGF0aCBkPSJNMC4wMzIyMjcsMzAuODhsLTAuMDMyMjI3LTE3LjA4NywyMy44NTMtMTMuNzkzLDIzLjc5NiwxMy43ODQtMTQuNjkxLDguNTEtOS4wNjItNS4xMDktMjMuODY0LDEzLjY5NXoiIGZpbGw9IiNmZGI5NDAiLz4KICAgIDxwYXRoIGQ9Ik0wLDQzLjM1NWwyMy44NzYsMTMuNjIyLDIzLjk3NC0xMy45Mzd2LTE2LjkwMmwtMjMuOTc0LDEzLjUwNi0yMy44NzYtMTMuNTA2djE3LjIxN3oiIGZpbGw9IiM1MjVjODYiLz4KICA8L2c+Cjwvc3ZnPgo=",
			},
			Plans: []brokerapi.ServicePlan{
				{
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
