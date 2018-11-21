package broker_client

import (
	"crypto/tls"
	"net/http"
)

type ServiceBrokerClient struct {
	HttpClient   *http.Client
	BrokerURL    string
	Token        string
	UserIdentity string
}

type BrokerClientClientConfig struct {
	BrokerURL    string
	Token        string
	UserIdentity string
	TlsCfg       *tls.Config
}

type Service struct {
	Name  string        `json:"name"`
	ID    string        `json:"id"`
	Plans []ServicePlan `json:"plans"`
}

type ServicePlan struct {
	ID string `json:"id"`
}

type ServiceBrokerError struct {
	Error       string `json:"error,omitempty"`
	Description string `json:"description,omitempty"`
}

type CatalogResponse struct {
	Services []*Service `json:"services"`
}

type ServiceInstanceResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

type DeleteResponse struct {
	Operation string `json:"operation,omitempty"`
}

type LastOperationResponse struct {
	State       string `json:"state"`
	Description string `json:"description,omitempty"`
}

type GetServiceInstanceResponse struct {
	ServiceId    string            `json:"service_id,omitempty"`
	PlanId       string            `json:"plan_id,omitempty"`
	DashboardURL string            `json:"dashboard_url,omitempty"`
	Parameters   map[string]string `json:"parameters,omitempty"`
}

type BindingResponse struct {
	Operation       string        `json:"operation,omitempty"`
	Credentials     interface{}   `json:"credentials,omitempty"`
	SyslogDrainURL  string        `json:"syslog_drain_url,omitempty"`
	VolumeMounts    []interface{} `json:"volume_mounts,omitempty"`
	RouteServiceURL string        `json:"route_service_url,omitempty"`
}

const (
	StateInProgress = "in progress"
	StateSucceeded  = "succeeded"
	StateFailed     = "failed"
)
