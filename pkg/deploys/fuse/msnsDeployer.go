package fuse

import (
	"fmt"
	apis "github.com/integr8ly/managed-service-broker/pkg/apis/integreatly/v1alpha1"
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	"github.com/integr8ly/managed-service-broker/pkg/deploys/fuse/pkg/apis/syndesis/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	glog "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"strings"
)

type MsnsFuseDeployer struct {
	k8sClient kubernetes.Interface
	osClient  *openshift.ClientFactory
}

func NewMsnsDeployer(k8sClient kubernetes.Interface, osClient *openshift.ClientFactory) *MsnsFuseDeployer {
	return &MsnsFuseDeployer{
		k8sClient: k8sClient,
		osClient: osClient,
	}
}

func (fd *MsnsFuseDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting fuse catalog entries")
	return getCatalogServicesObj()
}

func (fd *MsnsFuseDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	instanceId := req.InstanceId
	glog.Infof("Deploying fuse from deployer, id: %s", instanceId)

	msns := req.Msns
	fr, err := getFuseResource(msns.Name); if err != nil {
		return nil, err
	}

	if fr != nil {
		glog.Infof("Adding %s as a slice of fuse", instanceId)
		fr.Annotations[apis.SLICE_ANNOTATION + instanceId] = "true"
		if err := sdk.Update(fr); err != nil {
			return nil, errors.Wrap(err, "failed to add " + instanceId + "to the fuse custom resource")
		}
		return &brokerapi.ProvisionResponse{
			Code:         http.StatusAccepted,
			DashboardURL: "https://" + fr.Spec.RouteHostName,
		}, nil
	}

	frt := fd.createFuseCustomResourceTemplate(msns.Name, instanceId, msns.GetConsumerInstanceNamespace(instanceId), fd.k8sClient, req.OriginatingUserInfo.Username, req.Parameters)
	if err := sdk.Create(frt); err != nil {
		return nil, errors.Wrap(err, "failed to create a fuse custom resource")
	}

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: "https://" + frt.Spec.RouteHostName,
		Operation:    "deploy",
	}, nil
}

func (fd *MsnsFuseDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error){
	instanceId := req.InstanceId
	fr, err := getFuseResource(req.Msns.Name); if err != nil {
		return  &brokerapi.DeprovisionResponse{}, err
	}

	if fr == nil {
		return  &brokerapi.DeprovisionResponse{}, apierrors.NewNotFound(v1alpha1.SchemeGroupResource, instanceId)
	}

	key := apis.SLICE_ANNOTATION + instanceId
	_, ok := fr.Annotations[key]
	if !ok {
		return  &brokerapi.DeprovisionResponse{}, apierrors.NewNotFound(v1alpha1.SchemeGroupResource, instanceId)
	}

	delete(fr.Annotations, key)
	err = sdk.Update(fr); if err != nil {
		glog.Infof("Removing %s as a slice of fuse", instanceId)
		return  &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to remove instance id %s from fuse resource", instanceId))
	}

	if hasDeletedAllInstances(fr) {
		err = sdk.Delete(fr); if err != nil {
			return  &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to delete service instance: %s with error %+v", instanceId, err))
		}
	}

	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

// LastOperation should only return an error if it was unable to check the status, not if the status is failed, when status is failed, set that in the Response object
func (fd *MsnsFuseDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	instanceId := req.InstanceId
	glog.Infof("Getting last operation for %s", instanceId)

	fr, err := getFuseResource(req.Msns.Name); if err != nil {
		return nil, err
	}

	switch req.Operation {
	case "deploy":
		if fr == nil {
			return nil, apierrors.NewNotFound(v1alpha1.SchemeGroupResource, instanceId)
		}

		if fr.Status.Phase == v1alpha1.SyndesisPhaseStartupFailed {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateFailed,
				Description: fr.Status.Description,
			}, nil
		}

		if fr.Status.Phase == v1alpha1.SyndesisPhaseInstalled {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: fr.Status.Description,
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State: brokerapi.StateInProgress,
			Description: "fuse is deploying",
		}, nil

	case "remove":
		if fr == nil {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "Fuse has been deleted",
			}, nil
		}

		_, ok := fr.Annotations[apis.SLICE_ANNOTATION + instanceId]
		if !ok {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "Fuse has been deleted",
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State: brokerapi.StateInProgress,
			Description: "fr is deleting",
		}, nil
	default:
		if fr == nil {
			return nil, apierrors.NewNotFound(v1alpha1.SchemeGroupResource, instanceId)
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "unknown operation: " + req.Operation,
		}, nil
	}
}

func (fd *MsnsFuseDeployer) createFuseCustomResourceTemplate(managedNamespace, instanceID, consumerNamespace string, k8sclient kubernetes.Interface, userID string, parameters map[string]interface{}) *v1alpha1.Syndesis {
	integrationsLimit := 0
	if parameters["limit"] != nil {
		integrationsLimit = int(parameters["limit"].(float64))
	}

	fuseObj := getFuseObj(managedNamespace, consumerNamespace, integrationsLimit)
	fuseDashboardURL := fd.getRouteHostname(managedNamespace)
	fuseObj.Spec.RouteHostName = fuseDashboardURL
	fuseObj.Annotations["syndesis.io/created-by"] = userID
	fuseObj.Annotations[apis.SLICE_ANNOTATION + instanceID] = "true"

	return fuseObj
}

// Get route hostname for fuse
func (fd *MsnsFuseDeployer) getRouteHostname(namespace string) string {
	routeHostname := namespace
	routeSuffix, exists := os.LookupEnv("ROUTE_SUFFIX")
	if exists {
		routeHostname = routeHostname + "." + routeSuffix
	}
	return routeHostname
}

// Get fuse resource in namespace
func getFuseResource(ns string) (*v1alpha1.Syndesis, error) {
	syndList := v1alpha1.NewSyndesisList()
	err := sdk.List(ns, syndList); if err != nil {
		return nil, errors.Wrapf(err, "failed to list fuse resources with error %+v", err)
	}

	for _, synd := range syndList.Items {
		return &synd, nil
	}

	return nil, nil
}

// Check if all instances relying on the fuse resource have been deleted.
func hasDeletedAllInstances(synd *v1alpha1.Syndesis) bool {
	for k, _ := range synd.Annotations {
		if strings.HasPrefix(k, apis.SLICE_ANNOTATION) {
			return false
		}
	}

	return true
}