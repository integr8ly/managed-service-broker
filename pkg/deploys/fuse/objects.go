package fuse

import (
	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
	"github.com/aerogear/managed-services-broker/pkg/deploys/fuse/pkg/apis/syndesis/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	authv1 "github.com/openshift/api/authorization/v1"
	imagev1 "github.com/openshift/api/image/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Fuse plan
func getCatalogServicesObj() []*brokerapi.Service {
	return []*brokerapi.Service{
		{
			Name:        "fuse",
			ID:          "fuse-service-id",
			Description: "fuse",
			Metadata:    map[string]string{"serviceName": "fuse", "serviceType": "fuse"},
			Plans: []brokerapi.ServicePlan{
				brokerapi.ServicePlan{
					Name:        "default-fuse",
					ID:          "default-fuse",
					Description: "default fuse plan",
					Free:        true,
					Schemas: &brokerapi.Schemas{
						ServiceBinding: &brokerapi.ServiceBindingSchema{
							Create: &brokerapi.RequestResponseSchema{},
						},
						ServiceInstance: &brokerapi.ServiceInstanceSchema{
							Create: &brokerapi.InputParametersSchema{},
						},
					},
				},
			},
		},
	}
}

func getNamespaceObj(id string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: id,
		},
	}
}

// Fuse operator service account
func getServiceAccountObj() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
	}
}

// Fuse operator role
func getRoleObj() *rbacv1beta1.Role {
	return &rbacv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Rules: []rbacv1beta1.PolicyRule{
			{
				APIGroups: []string{"syndesis.io"},
				Resources: []string{"syndesises", "syndesises/finalizers"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "services", "endpoints", "persistentvolumeclaims", "configmaps", "secrets", "serviceaccounts"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"rolebindings"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"template.openshift.io"},
				Resources: []string{"processedtemplates"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"image.openshift.io"},
				Resources: []string{"imagestreams"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"apps.openshift.io"},
				Resources: []string{"deploymentconfigs"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"build.openshift.io"},
				Resources: []string{"buildconfigs"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"authorization.openshift.io"},
				Resources: []string{"rolebindings"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
			{
				APIGroups: []string{"route.openshift.io"},
				Resources: []string{"routes", "routes/custom-host"},
				Verbs:     []string{"get", "list", "create", "update", "delete", "deletecollection", "watch"},
			},
		},
	}
}

// Fuse specific role bindings
func getInstallRoleBindingObj() *rbacv1beta1.RoleBinding {
	return &rbacv1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator:install",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Subjects: []rbacv1beta1.Subject{
			{
				Kind: "ServiceAccount",
				Name: "syndesis-operator",
			},
		},
		RoleRef: rbacv1beta1.RoleRef{
			Kind:     "Role",
			Name:     "syndesis-operator",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func getViewRoleBindingObj() *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator:view",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "ServiceAccount",
				Name: "syndesis-operator",
			},
		},
		RoleRef: corev1.ObjectReference{
			Name: "view",
		},
	}
}

func getUserViewRoleBindingObj(namespace, username string) *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "syndesis-operator:view-",
			Namespace:    namespace,
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		RoleRef: corev1.ObjectReference{
			Name: "view",
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "User",
				Name: username,
			},
		},
	}
}

func getEditRoleBindingObj() *authv1.RoleBinding {
	return &authv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator:edit",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Subjects: []corev1.ObjectReference{
			{
				Kind: "ServiceAccount",
				Name: "syndesis-operator",
			},
		},
		RoleRef: corev1.ObjectReference{
			Name: "edit",
		},
	}
}

// Fuse image stream
func getImageStreamObj() *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Spec: imagev1.ImageStreamSpec{
			LookupPolicy: imagev1.ImageLookupPolicy{
				Local: true,
			},
			Tags: []imagev1.TagReference{
				{
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: "docker.io/jameelb/syndesis-operator:1.4.8", // NOTE: Point this to own version of syndesis-operator for auth
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: true,
					},
					Name: "fuse-7.1",
				},
			},
		},
	}
}

// Fuse operator deployment config
func getDeploymentConfigObj() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "syndesis-operator",
			Labels: map[string]string{
				"app":                   "syndesis",
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: "Recreate",
			},
			Replicas: 1,
			Selector: map[string]string{
				"syndesis.io/app":       "syndesis",
				"syndesis.io/type":      "operator",
				"syndesis.io/component": "syndesis-operator",
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"syndesis.io/app":       "syndesis",
						"syndesis.io/type":      "operator",
						"syndesis.io/component": "syndesis-operator",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "syndesis-operator",
					Containers: []corev1.Container{
						{
							Name:  "syndesis-operator",
							Image: " ",
							Command: []string{
								"syndesis-operator",
							},
							ImagePullPolicy: "IfNotPresent",
							Env: []corev1.EnvVar{
								{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				appsv1.DeploymentTriggerPolicy{
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic: true,
						ContainerNames: []string{
							"syndesis-operator",
						},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: "syndesis-operator:fuse-7.1",
						},
					},
					Type: "ImageChange",
				},
				appsv1.DeploymentTriggerPolicy{
					Type: "ConfigChange",
				},
			},
		},
	}
}

// System specific role bindings
func getSystemRoleBindings(namespace string) []rbacv1beta1.RoleBinding {
	return []rbacv1beta1.RoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:deployers",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "deployer",
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:deployer",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:image-builders",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "builder",
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:image-builder",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "system:image-pullers",
			},
			Subjects: []rbacv1beta1.Subject{
				{
					Kind:      "Group",
					Name:      "system:serviceaccounts:" + namespace,
					Namespace: namespace,
				},
			},
			RoleRef: rbacv1beta1.RoleRef{
				Kind:     "ClusterRole",
				Name:     "system:image-puller",
				APIGroup: "rbac.authorization.k8s.io",
			},
		},
	}
}

// Fuse Custom Resource
func getFuseObj(userNamespace string) *v1alpha1.Syndesis {
	demoData := false
	deployIntegrations := true
	limit := 1
	stateCheckInterval := 60

	return &v1alpha1.Syndesis{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Syndesis",
			APIVersion: "syndesis.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "fuse",
		},
		Spec: v1alpha1.SyndesisSpec{
			SarNamespace:         userNamespace,
			DemoData:             &demoData,
			DeployIntegrations:   &deployIntegrations,
			ImageStreamNamespace: "",
			Integration: v1alpha1.IntegrationSpec{
				Limit:              &limit,
				StateCheckInterval: &stateCheckInterval,
			},
			Registry: "docker.io",
			Components: v1alpha1.ComponentsSpec{
				Db: v1alpha1.DbConfiguration{
					Resources: v1alpha1.ResourcesWithVolume{
						ResourceRequirements: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": *resource.NewQuantity(255*1024*1024, resource.BinarySI),
							},
						},
						VolumeCapacity: "1Gi",
					},
					User:                 "syndesis",
					Database:             "syndesis",
					ImageStreamNamespace: "openshift",
				},
				Prometheus: v1alpha1.PrometheusConfiguration{
					Resources: v1alpha1.ResourcesWithVolume{
						ResourceRequirements: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": *resource.NewQuantity(512*1024*1024, resource.BinarySI),
							},
						},
						VolumeCapacity: "1Gi",
					},
				},
				Server: v1alpha1.ServerConfiguration{
					Resources: v1alpha1.Resources{
						ResourceRequirements: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": *resource.NewQuantity(800*1024*1024, resource.BinarySI),
							},
						},
					},
				},
				Meta: v1alpha1.MetaConfiguration{
					Resources: v1alpha1.ResourcesWithVolume{
						ResourceRequirements: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": *resource.NewQuantity(512*1024*1024, resource.BinarySI),
							},
						},
						VolumeCapacity: "1Gi",
					},
				},
			},
		},
	}
}
