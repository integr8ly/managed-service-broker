package fuse_managed

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type FuseManagedDeployer struct{}

func NewDeployer() *FuseManagedDeployer {
	return &FuseManagedDeployer{}
}

func (fd *FuseManagedDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting fuse managed catalog entries")
	return getCatalogServicesObj()
}

func (fd *FuseManagedDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying fuse managed from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("SHARED_FUSE_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
		Operation:    "deploy",
	}, nil
}

func (fd *FuseManagedDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (fd *FuseManagedDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "Managed Fuse deployed successfully",
	}, nil
}
