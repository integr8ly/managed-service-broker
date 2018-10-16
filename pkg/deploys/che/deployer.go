package che

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type CheDeployer struct {}

func NewDeployer() *CheDeployer {
	return &CheDeployer{}
}

func (cd *CheDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting che catalog entries")
	return getCatalogServicesObj()
}

func (cd *CheDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying che from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("CHE_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (cd *CheDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{}, nil
}

func (cd *CheDeployer) LastOperation (req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "che deployed successfully",
	}, nil
}
