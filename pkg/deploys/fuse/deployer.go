package fuse

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"k8s.io/api/authentication/v1"
	"github.com/integr8ly/managed-service-broker/pkg/deploys/fuse/pkg/apis/syndesis/v1alpha1"
	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	k8sClient "github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/pkg/errors"
	glog "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type FuseDeployer struct {
	id string
}

func NewDeployer(id string) *FuseDeployer {
	return &FuseDeployer{id: id}
}

func (fd *FuseDeployer) IsForService(serviceID string) bool {
	return serviceID == "fuse-service-id"
}

func (fd *FuseDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting fuse catalog entries")
	return getCatalogServicesObj()
}

func (fd *FuseDeployer) GetID() string {
	return fd.id
}

func (fd *FuseDeployer) Deploy(instanceID, namespace string, contextProfile brokerapi.ContextProfile, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) (*brokerapi.CreateServiceInstanceResponse, error) {
	glog.Infof("Deploying fuse from deployer, id: %s", instanceID)

	dashboardURL, err := fd.createFuseCustomResource(instanceID, namespace, contextProfile.Namespace, k8sclient, userInfo.Username)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.CreateServiceInstanceResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	return &brokerapi.CreateServiceInstanceResponse{
		Code:         http.StatusAccepted,
		DashboardURL: dashboardURL,
	}, nil
}

func (fd *FuseDeployer) RemoveDeploy(serviceInstanceId string, namespace string, k8sclient kubernetes.Interface) error {
	sds := v1alpha1.NewSyndesis()
	sds.Name = serviceInstanceId
	sds.Namespace = namespace
	err := sdk.Delete(sds);if err != nil {
		glog.Errorf("failed to delete service instance: %s with error %+v", serviceInstanceId, err)
		return errors.Wrap(err, fmt.Sprintf("failed to delete service instance: %s with error %+v", serviceInstanceId, err))
	}

	return nil
}


func (fd *FuseDeployer) LastOperation(instanceID, namespace string, k8sclient kubernetes.Interface, osclient *openshift.ClientFactory) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for %s", instanceID)
	podsToWatch := []string{"syndesis-oauthproxy", "syndesis-server", "syndesis-ui"}

	dcClient, err := osclient.AppsClient()
	if err != nil {
		glog.Errorf("failed to create an openshift deployment config client: %+v", err)
		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "Failed to create an openshift deployment config client",
		}, errors.Wrap(err, "failed to create an openshift deployment config client")
	}

	for _, v := range podsToWatch {
		state, description, err := fd.getPodStatus(v, namespace, dcClient)
		if state != brokerapi.StateSucceeded {
			return &brokerapi.LastOperationResponse{
				State:       state,
				Description: description,
			}, err
		}
	}

	return &brokerapi.LastOperationResponse{
		State:       brokerapi.StateSucceeded,
		Description: "fuse deployed successfully",
	}, nil
}

func (fd *FuseDeployer) createFuseCustomResource(instanceID, managedNamespace, userNamespace string, k8sclient kubernetes.Interface, userID string) (string, error) {
	fuseClient, _, err := k8sClient.GetResourceClient("syndesis.io/v1alpha1", "Syndesis", managedNamespace)
	if err != nil {
		return "", errors.Wrap(err, "failed to create fuse client")
	}

	fuseObj := getFuseObj(instanceID, userNamespace)
	fuseObj.Annotations["syndesis.io/created-by"] = userID

	fuseDashboardURL := fd.getRouteHostname(managedNamespace)

	fuseObj.Spec.RouteHostName = fuseDashboardURL
	_, err = fuseClient.Create(k8sutil.UnstructuredFromRuntimeObject(fuseObj))
	if err != nil {
		return "", errors.Wrap(err, "failed to create a fuse custom resource")
	}

	return "https://" + fuseDashboardURL, nil
}

// Get route hostname for fuse
func (fd *FuseDeployer) getRouteHostname(namespace string) string {
	routeHostname := namespace
	routeSuffix, exists := os.LookupEnv("ROUTE_SUFFIX")
	if exists {
		routeHostname = routeHostname + "." + routeSuffix
	}
	return routeHostname
}

func (fd *FuseDeployer) getPodStatus(podName, namespace string, dcClient *appsv1.AppsV1Client) (string, string, error) {
	pod, err := dcClient.DeploymentConfigs(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("Failed to get status of %s: %+v", podName, err)
		return brokerapi.StateFailed,
			"Failed to get status of " + podName,
			errors.Wrap(err, "failed to get status of "+podName)
	}

	for _, v := range pod.Status.Conditions {
		if v.Status == "False" {
			return brokerapi.StateInProgress, v.Message, nil
		}
	}

	return brokerapi.StateSucceeded, "", nil
}
