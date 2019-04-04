package sso

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type RHSSOManagedDeployer struct{}

func NewDeployer() *RHSSOManagedDeployer {
	return &RHSSOManagedDeployer{}
}

func (fd *RHSSOManagedDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting RH-SSO managed catalog entries")
	return getCatalogServicesObj()
}

func (fd *RHSSOManagedDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying RH-SSO from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("SSO_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
		Operation:    "deploy",
	}, nil
}

func (fd *RHSSOManagedDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (fd *RHSSOManagedDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "RH-SSO deployed successfully",
	}, nil
}
