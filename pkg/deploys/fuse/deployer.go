package fuse

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
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
	yamlv2 "gopkg.in/yaml.v2"
	"k8s.io/api/authentication/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	k8errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fusePullSecretName = "imagestreamsecret"
	fusePullSecretKey  = ".dockerconfigjson"
)

type FuseDeployer struct {
	k8sClient     kubernetes.Interface
	osClient      *openshift.ClientFactory
	client        client.Client
	monitoringKey string
}

func NewDeployer(k8sClient kubernetes.Interface, osClient *openshift.ClientFactory, client client.Client, mk string) *FuseDeployer {
	return &FuseDeployer{
		k8sClient:     k8sClient,
		osClient:      osClient,
		client:        client,
		monitoringKey: mk,
	}
}

func (fd *FuseDeployer) GetCatalogEntries() []*brokerapi.Service {
	glog.Infof("Getting fuse catalog entries")
	return getCatalogServicesObj()
}

func (fd *FuseDeployer) Deploy(req *brokerapi.ProvisionRequest, async bool) (*brokerapi.ProvisionResponse, error) {
	glog.Infof("Deploying fuse from deployer, id: %s", req.InstanceId)

	// Namespace
	ns, err := fd.k8sClient.CoreV1().Namespaces().Create(getNamespaceObj("fuse-"+req.InstanceId, fd.monitoringKey))
	if err != nil {
		glog.Errorf("failed to create fuse namespace: %+v", err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, errors.Wrap(err, "failed to create namespace for fuse service")
	}

	namespace := ns.ObjectMeta.Name

	err = fd.createImagePullSecret(namespace)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	err = fd.createOperatorResources(namespace, fd.client)
	if err != nil {
		glog.Errorln(err)
		return &brokerapi.ProvisionResponse{
			Code: http.StatusInternalServerError,
		}, err
	}

	//RoleBindings
	err = fd.createRoleBindings(namespace, req.OriginatingUserInfo, fd.k8sClient, fd.osClient)
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

// Creates the syndesis pull secret required to pull images from registry.redhat.io
func (fd *FuseDeployer) createImagePullSecret(userNamespace string) error {
	operatorNamespace := os.Getenv("POD_NAMESPACE")
	if operatorNamespace == "" {
		return errors.New("POD_NAMESPACE must be set")
	}

	pullSecret, err := fd.k8sClient.CoreV1().Secrets(operatorNamespace).Get(fusePullSecretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, ok := pullSecret.Data[fusePullSecretKey]; !ok {
		return errors.New(fmt.Sprintf("%s does not contain a key named %s", fusePullSecretName, fusePullSecretKey))
	}

	_, err = fd.k8sClient.CoreV1().Secrets(userNamespace).Create(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: userNamespace,
			Name:      fusePullSecretName,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: pullSecret.Data,
	})

	return err
}

func (fd *FuseDeployer) createOperatorResources(namespace string, client client.Client) error {
	resourcesUrl, exists := os.LookupEnv("FUSE_OPERATOR_RESOURCES_URL")
	if !exists {
		return errors.New("FUSE_OPERATOR_RESOURCES_URL environment variable is not set")
	}
	glog.Printf("Operator resources url = %v", resourcesUrl)

	var httpClient http.Client
	resp, err := httpClient.Get(resourcesUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var resources []runtime.Object
	dec := yamlv2.NewDecoder(resp.Body)
	for {
		var value interface{}
		err := dec.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		yamlData, err := yamlv2.Marshal(value)
		if err != nil {
			return err
		}
		jsonData, err := yaml.ToJSON(yamlData)
		if err != nil {
			return err
		}
		resource, err := openshift.LoadKubernetesResource(jsonData, namespace)
		if err != nil {
			return err
		}
		resources = append(resources, resource)
	}
	//ToDo Can we lazy load these resources so we don't need to be doing a http request every time

	for _, resource := range resources {
		err = client.Create(context.TODO(), resource)
		if err != nil && !k8errors.IsAlreadyExists(err) {
			glog.Errorf("failed to create object during provision with kind %v, err: %+v", resource.GetObjectKind().GroupVersionKind().String(), err)
			return err
		}
	}
	return nil
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

	authClient, err := osClientFactory.AuthClient()
	if err != nil {
		return errors.Wrap(err, "failed to create an openshift authorization client")
	}

	_, err = authClient.RoleBindings(namespace).Create(getUserViewRoleBindingObj(namespace, userInfo.Username))
	if err != nil {
		return errors.Wrap(err, "failed to create user view role binding for fuse service")
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
