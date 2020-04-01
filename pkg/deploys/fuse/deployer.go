package fuse

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	synv1 "github.com/integr8ly/managed-service-broker/pkg/apis/syndesis/v1beta1"
	v1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	brokerapi "github.com/integr8ly/managed-service-broker/pkg/broker"
	"github.com/integr8ly/managed-service-broker/pkg/clients/openshift"
	k8sClient "github.com/operator-framework/operator-sdk/pkg/k8sclient"
	"github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	"github.com/pkg/errors"
	glog "github.com/sirupsen/logrus"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fusePullSecretName = "syndesis-pull-secret"
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
		DashboardURL: "https://" + frt.Spec.RouteHostname,
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

	destUrl := "/tmp"
	destBin := fmt.Sprintf("%s/syndesis-operator", destUrl)
	if _, err := os.Stat(destBin); os.IsNotExist(err) || err != nil {
		glog.Infof("downloading fuse online binary from %s, to %s", resourcesUrl, destUrl)
		if err := downloadSyndesisBinary(resourcesUrl, destUrl); err != nil {
			glog.Infof("failed to download fuse online binary")
			return err
		}
	}

	if output, err := exec.Command(destBin, "install", "operator", "--wait", "-n", namespace).Output(); err != nil {
		if output != nil {
			glog.Infof("Output: %s", string(output))
		}
		glog.Infof("failed to install fuse online operator in namespace %s", namespace)
		return err
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
			return nil, apiErrors.NewNotFound(schema.GroupResource{
				Group:    synv1.SchemeGroupVersion.Group,
				Resource: synv1.SchemaGroupVersionKind.Kind,
			}, req.InstanceId)
		}

		if fr.Status.Phase == synv1.SyndesisPhaseStartupFailed {
			return &brokerapi.LastOperationResponse{
				State:       brokerapi.StateFailed,
				Description: fr.Status.Description,
			}, nil
		}

		if fr.Status.Phase == synv1.SyndesisPhaseInstalled {
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
			return nil, apiErrors.NewNotFound(schema.GroupResource{Group: synv1.SchemeGroupVersion.Group, Resource: "Syndesis"}, req.InstanceId)
		}

		return &brokerapi.LastOperationResponse{
			State:       brokerapi.StateFailed,
			Description: "unknown operation: " + req.Operation,
		}, nil
	}
}

func (fd *FuseDeployer) createRoleBindings(namespace string, userInfo v1.UserInfo, k8sclient kubernetes.Interface, osClientFactory *openshift.ClientFactory) error {
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
func (fd *FuseDeployer) createFuseCustomResourceTemplate(namespace, userNamespace, userID string, parameters map[string]interface{}) *synv1.Syndesis {
	integrationsLimit := 0
	if parameters["limit"] != nil {
		integrationsLimit = int(parameters["limit"].(float64))
	}

	fuseObj := getFuseObj(namespace, userNamespace, integrationsLimit)
	fuseDashboardURL := fd.getRouteHostname(namespace)
	fuseObj.Spec.RouteHostname = fuseDashboardURL
	fuseObj.Annotations["syndesis.io/created-by"] = userID

	// Handle exposing via 3scale if management URL is set for 3scale
	threescaleDashboardUrl := os.Getenv("THREESCALE_DASHBOARD_URL")
	if threescaleDashboardUrl != "" {
		fuseObj.Spec.Components.Server.Features.ManagementUrlFor3scale = threescaleDashboardUrl
	}
	return fuseObj
}

// Create the fuse custom resource
func (fd *FuseDeployer) createFuseCustomResource(namespace string, fr *synv1.Syndesis) error {
	fuseClient, _, err := k8sClient.GetResourceClient(synv1.SchemeGroupVersion.Group+"/"+synv1.SchemeGroupVersion.Version, synv1.SchemaGroupVersionKind.Kind, namespace)
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
func getFuse(ns string) (*synv1.Syndesis, error) {
	fuseClient, _, err := k8sClient.GetResourceClient(synv1.SchemeGroupVersion.Group+"/"+synv1.SchemeGroupVersion.Version, synv1.SchemaGroupVersionKind.Kind, ns)
	if err != nil {
		return nil, err
	}

	u, err := fuseClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	fl := &synv1.SyndesisList{}
	if err := k8sutil.RuntimeObjectIntoRuntimeObject(u, fl); err != nil {
		return nil, errors.Wrap(err, "failed to get the fuse resources")
	}

	for _, f := range fl.Items {
		return &f, nil
	}

	return nil, nil
}

func downloadSyndesisBinary(srcUrl, destUrl string) error {
	resp, err := http.Get(srcUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := untar(destUrl, resp.Body); err != nil {
		return err
	}
	if err := os.Chmod(fmt.Sprintf("%s/syndesis-operator", destUrl), 755); err != nil {
		return err
	}
	return nil
}

func untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
