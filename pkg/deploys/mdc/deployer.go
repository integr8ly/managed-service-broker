package mdc

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type MDCDeployer struct {
	id string
}

func NewDeployer() *MDCDeployer {
	return &MDCDeployer{}
}

func (md *MDCDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting mdc catalog entries")
	return getCatalogServicesObj()
}

func (md *MDCDeployer) GetID() string {
	return md.id
}

func (md *MDCDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying mdc from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("MDC_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (md *MDCDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (ac *MDCDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "mdc deployed successfully",
	}, nil
}
