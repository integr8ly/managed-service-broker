package threescale

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type ThreeScaleDeployer struct {}

func NewDeployer() *ThreeScaleDeployer {
	return &ThreeScaleDeployer{}
}

func (fd *ThreeScaleDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting 3scale catalog entries")
	return getCatalogServicesObj()
}

func (fd *ThreeScaleDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying 3scale from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("THREESCALE_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
		Operation:    "deploy",
	}, nil
}

func (fd *ThreeScaleDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return  &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (fd *ThreeScaleDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "3scale deployed successfully",
	}, nil
}
