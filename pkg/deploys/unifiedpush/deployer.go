package unifiedpush

import (
	"net/http"
	"os"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	glog "github.com/sirupsen/logrus"
)

type UnifiedpushDeployer struct {
	id string
}

func NewDeployer() *UnifiedpushDeployer {
	return &UnifiedpushDeployer{}
}

func (md *UnifiedpushDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting unifiedpush catalog entries")
	return getCatalogServicesObj()
}

func (md *UnifiedpushDeployer) GetID() string {
	return md.id
}

func (md *UnifiedpushDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying unifiedpush from deployer, id: %s", req.InstanceId)

	dashboardUrl := os.Getenv("UNIFIEDPUSH_DASHBOARD_URL")

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardUrl,
	}, nil
}

func (md *UnifiedpushDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

func (ac *UnifiedpushDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", req.InstanceId)

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "unifiedpush deployed successfully",
	}, nil
}
