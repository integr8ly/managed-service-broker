package openshift

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"

	"encoding/json"
	apps "github.com/openshift/api/apps/v1"
	authorization "github.com/openshift/api/authorization/v1"
	build "github.com/openshift/api/build/v1"
	image "github.com/openshift/api/image/v1"
	route "github.com/openshift/api/route/v1"
	template "github.com/openshift/api/template/v1"
)

var (
	AddToSchemes runtime.SchemeBuilder
	Scheme       = scheme.Scheme
	codecs       = serializer.NewCodecFactory(Scheme)
	decoderFunc  = decoder
)

func init() {
	AddToSchemes = append(AddToSchemes,
		apps.Install,
		authorization.Install,
		build.Install,
		image.Install,
		route.Install,
		template.Install,
	)
}

func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}

func decoder(gv schema.GroupVersion, codecs serializer.CodecFactory) runtime.Decoder {
	codec := codecs.UniversalDecoder(gv)
	return codec
}

func LoadKubernetesResource(jsonData []byte, namespace string) (runtime.Object, error) {
	u := unstructured.Unstructured{}

	err := u.UnmarshalJSON(jsonData)
	if err != nil {
		return nil, err
	}
	u.SetNamespace(namespace)
	//ToDo Is there a way to register the DeploymentConfig without the Group?
	if u.GetObjectKind().GroupVersionKind().Kind == "DeploymentConfig" {
		u.GetObjectKind().SetGroupVersionKind(
			schema.GroupVersionKind{
				Version: "v1",
				Group:   "apps.openshift.io",
				Kind:    "DeploymentConfig",
			})
	}
	//
	return runtimeObjectFromUnstructured(&u)
}

func runtimeObjectFromUnstructured(u *unstructured.Unstructured) (runtime.Object, error) {
	gvk := u.GroupVersionKind()
	decoder := decoderFunc(gvk.GroupVersion(), codecs)

	b, err := u.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on unstructured object: %v", err)
	}

	ro, _, err := decoder.Decode(b, &gvk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json data with gvk(%v): %v", gvk.String(), err)
	}

	return ro, nil
}

func UnstructuredFromRuntimeObject(ro runtime.Object) (*unstructured.Unstructured, error) {
	b, err := json.Marshal(ro)
	if err != nil {
		return nil, fmt.Errorf("error running MarshalJSON on runtime object: %v", err)
	}
	var u unstructured.Unstructured
	if err := json.Unmarshal(b, &u.Object); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json into unstructured object: %v", err)
	}
	return &u, nil
}
