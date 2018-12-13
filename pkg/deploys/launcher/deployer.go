package launcher

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type LauncherDeployer struct{}

func NewDeployer() *LauncherDeployer {
	return &LauncherDeployer{}
}

func (ld *LauncherDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting launcher catalog entries")
	return getCatalogServicesObj()
}

func (ld *LauncherDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying launcher from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("LAUNCHER_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
		Operation:    "deploy",
	}, nil
}

func (ld *LauncherDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (ld *LauncherDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "launcher deployed successfully",
	}, nil
}
