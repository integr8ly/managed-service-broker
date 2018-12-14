package fuse

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	fuseV1alpha1 "github.com/integr8ly/managed-service-broker/pkg/deploys/fuse/pkg/apis/syndesis/v1alpha1"
	k8sClient "github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/pkg/errors"
	glog "github.com/sirupsen/logrus"
	"k8s.io/api/authentication/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type FuseDeployer struct {
	k8sClient kubernetes.Interface
	osClient  *openshift.ClientFactory
}

func NewDeployer(k8sClient kubernetes.Interface, osClient *openshift.ClientFactory) *FuseDeployer {
	return &FuseDeployer{
		k8sClient: k8sClient,
		osClient:  osClient,
	}
}

func (fd *FuseDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting fuse catalog entries")
	return getCatalogServicesObj()
}

func (fd *FuseDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying fuse from deployer, id: %s", req.InstanceId)

	// Namespace
	ns, err := fd.k8sClient.CoreV1().Namespaces().Create(getNamespaceObj("fuse-" + req.InstanceId))
	if err != nil {
		glog.Errorf("failed to create fuse namespace: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create namespace for fuse service")
	}

	namespace := ns.ObjectMeta.Name

	// ServiceAccount
	_, err = fd.k8sClient.CoreV1().ServiceAccounts(namespace).Create(getServiceAccountObj())
	if err != nil {
		glog.Errorf("failed to create fuse service account: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create service account for fuse service")
	}

	//Role
	_, err = fd.k8sClient.RbacV1beta1().Roles(namespace).Create(getRoleObj())
	if err != nil {
		glog.Errorf("failed to create fuse role: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create role for fuse service")
	}

	// RoleBindings
	err = fd.createRoleBindings(namespace, req.OriginatingUserInfo, fd.k8sClient, fd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// DeploymentConfig
	err = fd.createFuseOperator(namespace, fd.osClient)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	// Fuse custom resource
	frt := fd.createFuseCustomResourceTemplate(namespace, req.ContextProfile.Namespace, req.OriginatingUserInfo.Username, req.Parameters)
	if err := fd.createFuseCustomResource(namespace, frt); err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	return &brokerapi.ProvisionResponse{
		Code:         http.StatusAccepted,
		DashboardURL: "https://" + frt.Spec.RouteHostName,
		Operation:    "deploy",
	}, nil
}

func (fd *FuseDeployer) RemoveDeploy(req *brokerapi.DeprovisionRequest, async bool) (*brokerapi.DeprovisionResponse, error) {
	ns := "fuse-" + req.InstanceId
	err := fd.k8sClient.CoreV1().Namespaces().Delete(ns, &metav1.DeleteOptions{})
	if err != nil && !strings.Contains(err.Error(), "not found") {
		glog.Errorf("failed to delete %s namespace: %+v", ns, err)
		return &brokerapi.DeprovisionResponse{}, errors.Wrap(err, fmt.Sprintf("failed to delete namespace %s", ns))
	} else if err != nil && strings.Contains(err.Error(), "not found") {
		glog.Infof("fuse namespace already deleted")
	}
	return &brokerapi.DeprovisionResponse{Operation: "remove"}, nil
}

// ServiceInstanceLastOperation should only return an error if it was unable to check the status, not if the status is failed, when status is failed, set that in the Response object
func (fd *FuseDeployer) ServiceInstanceLastOperation(req *brokerapi.LastOperationRequest) (*brokerapi.LastOperationResponse, error) {
	glog.Infof("Getting last operation for service %s", req.InstanceId)

	namespace := "fuse-" + req.InstanceId
	switch req.Operation {
	case "deploy":
		fr, err := getFuse(namespace)
		if err != nil {
			return nil, err
		}
		if fr == nil {
			return nil, apiErrors.NewNotFound(fuseV1alpha1.SchemeGroupResource, req.InstanceId)
		}

		if fr.Status.Phase == fuseV1alpha1.SyndesisPhaseStartupFailed {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateFailed,
				Description: fr.Status.Description,
			}, nil
		}

		if fr.Status.Phase == fuseV1alpha1.SyndesisPhaseInstalled {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: fr.Status.Description,
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateInProgress,
			Description: "fuse is deploying",
		}, nil

	case "remove":
		_, err := fd.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
		if err != nil && apiErrors.IsNotFound(err) {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateSucceeded,
				Description: "Fuse has been deleted",
			}, nil
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateInProgress,
			Description: "fr is deleting",
		}, nil
	default:
		fr, err := getFuse(namespace)
		if err != nil {
			return nil, err
		}
		if fr == nil {
			return nil, apiErrors.NewNotFound(fuseV1alpha1.SchemeGroupResource, req.InstanceId)
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "unknown operation: " + req.Operation,
		}, nil
	}
}

func (fd *FuseDeployer) createRoleBindings(namespace string, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) error {
	for _, sysRoleBinding := range getSystemRoleBindings(namespace) {
		_, err := k8sclient.RbacV1beta1().RoleBindings(namespace).Create(&sysRoleBinding)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return errors.Wrapf(err, "failed to create rolebinding for %s", &sysRoleBinding.ObjectMeta.Name)
		}
	}

	_, err := k8sclient.RbacV1beta1().RoleBindings(namespace).Create(getInstallRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create install role binding for fuse service")
	}

	authClient, err := osClientFactory.AuthClient()
	if err != nil {
		return errors.Wrap(err, "failed to create an openshift authorization client")
	}

	_, err = authClient.RoleBindings(namespace).Create(getViewRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create view role binding for fuse service")
	}

	_, err = authClient.RoleBindings(namespace).Create(getEditRoleBindingObj())
	if err != nil {
		return errors.Wrap(err, "failed to create edit role binding for fuse service")
	}

	_, err = authClient.RoleBindings(namespace).Create(getUserViewRoleBindingObj(namespace, userInfo.Username))
	if err != nil {
		return errors.Wrap(err, "failed to create user view role binding for fuse service")
	}

	return nil
}

func (fd *FuseDeployer) createFuseOperator(namespace string, osClientFactory *openshift.ClientFactory) error {
	dcClient, err := osClientFactory.AppsClient()
	if err != nil {
		return errors.Wrap(err, "failed to create an openshift deployment config client")
	}

	_, err = dcClient.DeploymentConfigs(namespace).Create(getDeploymentConfigObj())
	if err != nil {
		return errors.Wrap(err, "failed to create deployment config for fuse service")
	}

	return nil
}

// Create the fuse custom resource template
func (fd *FuseDeployer) createFuseCustomResourceTemplate(namespace, userNamespace, userID string, parameters map[string]interface{}) *fuseV1alpha1.Syndesis {
	integrationsLimit := 0
	if parameters["limit"] != nil {
		integrationsLimit = int(parameters["limit"].(float64))
	}

	fuseObj := getFuseObj(namespace, userNamespace, integrationsLimit)
	fuseDashboardURL := fd.getRouteHostname(namespace)
	fuseObj.Spec.RouteHostName = fuseDashboardURL
	fuseObj.Annotations["syndesis.io/created-by"] = userID

	return fuseObj
}

// Create the fuse custom resource
func (fd *FuseDeployer) createFuseCustomResource(namespace string, fr *fuseV1alpha1.Syndesis) error {
	fuseClient, _, err := k8sClient.GetResourceClient("syndesis.io/v1alpha1", "Syndesis", namespace)
	if err != nil {
		return err
	}

	_, err = fuseClient.Create(k8sutil.UnstructuredFromRuntimeObject(fr))
	if err != nil {
		return err
	}

	return nil
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

// Get fuse resource in namespace
func getFuse(ns string) (*fuseV1alpha1.Syndesis, error) {
	fuseClient, _, err := k8sClient.GetResourceClient("syndesis.io/v1alpha1", "Syndesis", ns)
	if err != nil {
		return nil, err
	}

	u, err := fuseClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	fl := fuseV1alpha1.NewSyndesisList()
	if err := k8sutil.RuntimeObjectIntoRuntimeObject(u, fl); err != nil {
		return nil, errors.Wrap(err, "failed to get the fuse resources")
	}

	for _, f := range fl.Items {
		return &f, nil
	}

	return nil, nil
}
