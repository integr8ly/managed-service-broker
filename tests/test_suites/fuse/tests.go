package fuse

import (
	"fmt"
	integreatly "github.com/integr8ly/managed-service-broker/pkg/apis/integreatly/v1alpha1"
	sndv1alpha1 "github.com/integr8ly/managed-service-broker/pkg/deploys/fuse/pkg/apis/syndesis/v1alpha1"
	testapi "github.com/integr8ly/managed-service-broker/tests/apis"
	brokerClient "github.com/integr8ly/managed-service-broker/tests/broker_client"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"testing"
)

func TestSharedFuse(t *testing.T, tc *testapi.TestCase, svcp *brokerClient.ServicePlan, managedNamespace string, sbc *brokerClient.ServiceBrokerClient) {
	svc1 := "fuse1"
	svc2 := "fuse2"
	svc1SliceAnnotation := integreatly.SLICE_ANNOTATION + svc1

	_, csiRes1, err := sbc.CreateServiceInstance(tc.Namespace, svc1, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error creating %s", svc1), err.Description)
	}
	// We need to wait for first Provision to succeed as the resource is being updated and makes any following updates fail on occasion.
	sbc.PollLastOperation(svc1, tc.Service, svcp, csiRes1.Operation)

	_, csiRes2, err := sbc.CreateServiceInstance(tc.Namespace, svc2, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("Unexpected error creating %s", svc2), err.Description)
	}

	if csiRes1.DashboardURL != csiRes2.DashboardURL {
		t.Fatal("Dashboard URLs were not the same")
	}

	// Assert that one custom resource exists in the Managed Namespace and it has the correct 2 annotations
	fL := getFuseList(managedNamespace)
	expectedNumFuseCustomResources := 1
	if len(fL.Items) != expectedNumFuseCustomResources {
		t.Fatal("There should be one fuse custom resource")
	}
	f := fL.Items[0]
	expectedAnnotations := []string{svc1SliceAnnotation, integreatly.SLICE_ANNOTATION + svc2}
	for _, v := range expectedAnnotations {
		_, ok := f.Annotations[v]
		if !ok {
			t.Fatal(fmt.Sprintf("There should be an annotation of %s on the fuse custom resource", v))
		}
	}

	_, dRes1, err := sbc.DeleteServiceInstance(svc1, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("An error has occured deleting service instance %s", svc1), err.Description)
	}
	sbc.PollLastOperation(svc1, tc.Service, svcp, dRes1.Operation)

	// Assert that custom resource still exists and now has only 1 annotation
	fL2 := sndv1alpha1.NewSyndesisList()
	sdk.List(managedNamespace, fL2)
	if len(fL2.Items) != expectedNumFuseCustomResources {
		t.Fatal("There should be one fuse custom resource")
	}
	f = fL2.Items[0]
	_, ok := f.Annotations[svc1SliceAnnotation]
	if ok {
		t.Fatal(fmt.Sprintf("The annotation %s should be removed from the fuse custom resource", svc1SliceAnnotation))
	}

	_, dRes2, err := sbc.DeleteServiceInstance(svc2, tc.Service.ID, svcp.ID, tc.Async)
	if err != nil {
		t.Fatal(fmt.Sprintf("An error has occured deleting service instance %s", svc2), err.Description)
	}
	sbc.PollLastOperation(svc2, tc.Service, svcp, dRes2.Operation)

	// Assert that custom resource is deleted
	fl3 := getFuseList(managedNamespace)
	if len(fl3.Items) != 0 {
		t.Fatal("The fuse custom resource should be deleted")
	}
}

func getFuseList(ns string) *sndv1alpha1.SyndesisList {
	fL := sndv1alpha1.NewSyndesisList()
	sdk.List(ns, fL)
	return fL
}
