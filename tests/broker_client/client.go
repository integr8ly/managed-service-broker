package broker_client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

func (sbc *ServiceBrokerClient) Catalog() (*http.Response, *CatalogResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/catalog", sbc.BrokerURL)
	req, err := createRequest(http.MethodGet, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &CatalogResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*CatalogResponse), nil
}

func (sbc *ServiceBrokerClient) CreateServiceInstance(
	namespace,

	instanceId,
	serviceId,
	planId string,
	acceptsIncomplete bool) (*http.Response, *ServiceInstanceResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s?accepts_incomplete=%t", sbc.BrokerURL, instanceId, acceptsIncomplete)
	data := `{
                 "plan_id": %q,
                 "service_id": %q,
                 "context": {
                     "platform":"kubernetes",
                     "namespace": %q
                 },
                 "parameters": {}
              }`
	req, err := createRequest(http.MethodPut, url, fmt.Sprintf(data, planId, serviceId, namespace), sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &ServiceInstanceResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*ServiceInstanceResponse), nil
}

func (sbc *ServiceBrokerClient) DeleteServiceInstance(
	instanceId,
	serviceId,
	planId string,
	acceptsIncomplete bool) (*http.Response, *DeleteResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s?service_id=%s&plan_id=%s&accepts_incomplete=%t", sbc.BrokerURL, instanceId, serviceId, planId, acceptsIncomplete)
	req, err := createRequest(http.MethodDelete, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &DeleteResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*DeleteResponse), nil
}

func (sbc *ServiceBrokerClient) GetServiceInstanceLastOperation(
	instanceId,
	serviceId,
	planId,
	operation string) (*http.Response, *LastOperationResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s/last_operation?operation=%s&service_id=%s&plan_id=%s", sbc.BrokerURL, instanceId, operation, serviceId, planId)
	req, err := createRequest(http.MethodGet, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &LastOperationResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*LastOperationResponse), nil
}

func (sbc *ServiceBrokerClient) PollLastOperation(instanceId string, svc *Service, svcp *ServicePlan, operation string) error {
	if len(operation) == 0 || len(instanceId) == 0 {
		return errors.New("instanceId and operation are required.")
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		hasFinished := false
		for hasFinished == false {
			// Ignore errors/error states as GetServiceInstanceLastOperation can return errors/error states as operations are initialising.
			_, loRes, _ := sbc.GetServiceInstanceLastOperation(instanceId, svc.ID, svcp.ID, operation)

			time.Sleep(time.Second * 1)
			// Continue polling until the Service Broker returns a valid response.
			hasFinished = loRes != nil && loRes.State == StateSucceeded
		}
	}()
	wg.Wait()

	return nil
}

func (sbc *ServiceBrokerClient) GetServiceBindingLastOperation(
	instanceId,
	bindingId,
	serviceId,
	planId,
	operation string) (*http.Response, *LastOperationResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s/last_operation?operation=%s&service_id=%s&plan_id=%s", sbc.BrokerURL, instanceId, bindingId, operation, serviceId, planId)
	req, err := createRequest(http.MethodGet, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &LastOperationResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*LastOperationResponse), nil
}

func (sbc *ServiceBrokerClient) GetServiceInstance(instanceId string) (*http.Response, *GetServiceInstanceResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s", sbc.BrokerURL, instanceId)
	req, err := createRequest(http.MethodGet, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &GetServiceInstanceResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*GetServiceInstanceResponse), nil
}

func (sbc *ServiceBrokerClient) UpdateServiceInstance(
	instanceId,
	serviceId,
	planId string,
	acceptsIncomplete bool) (*http.Response, *ServiceInstanceResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s?accepts_incomplete=%t", sbc.BrokerURL, instanceId, acceptsIncomplete)
	data := `{
                 "plan_id": %q,
                 "service_id": %q,
                 "context": {
                     "platform":"kubernetes"
                 },
                 "previous_values":{},
                 "parameters": {}
              }`
	req, err := createRequest(http.MethodPatch, url, fmt.Sprintf(data, planId, serviceId), sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &ServiceInstanceResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*ServiceInstanceResponse), nil
}

func (sbc *ServiceBrokerClient) CreateServiceInstanceBinding(
	instanceId,
	serviceId,
	planId string,
	bindingId,
	acceptsIncomplete bool) (*http.Response, *BindingResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s", sbc.BrokerURL, instanceId, bindingId, acceptsIncomplete)
	data := `{
                 "plan_id": %q,
                 "service_id": %q,
                 "context": {
                     "platform":"kubernetes",
                     "namespace": %q
                 },
                 "bind_resource": {},
                 "parameters": {}
              }`
	req, err := createRequest(http.MethodPut, url, fmt.Sprintf(data, planId, serviceId), sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &BindingResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*BindingResponse), nil
}

func (sbc *ServiceBrokerClient) GetServiceInstanceBinding(instanceId, bindingId string) (*http.Response, *BindingResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%s/service_bindings/%s", sbc.BrokerURL, instanceId, bindingId)
	req, err := createRequest(http.MethodGet, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}
	res, obj, err := doRequest(req, sbc.HttpClient, &BindingResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*BindingResponse), nil
}

func (sbc *ServiceBrokerClient) DeleteServiceInstanceBinding(
	instanceId,
	serviceId,
	planId,
	bindingId string,
	acceptsIncomplete bool) (*http.Response, *DeleteResponse, *ServiceBrokerError) {
	url := fmt.Sprintf("%s/v2/service_instances/%sservice_bindings/%s?service_id=%s&plan_id=%s&accepts_incomplete=%t", sbc.BrokerURL, instanceId, bindingId, serviceId, planId, acceptsIncomplete)
	req, err := createRequest(http.MethodDelete, url, "", sbc.Token, sbc.UserIdentity)
	if err != nil {
		return nil, nil, err
	}

	res, obj, err := doRequest(req, sbc.HttpClient, &DeleteResponse{})
	if err != nil {
		return res, nil, err
	}

	return res, obj.(*DeleteResponse), nil
}

func createRequest(method, url, data, token, userIdentity string) (*http.Request, *ServiceBrokerError) {
	req, err := http.NewRequest(method, url, bytes.NewBufferString(data))
	if err != nil {
		return nil, NewServiceBrokerError(err)
	}

	req.Header.Add("Content-type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("bearer %s", token))
	req.Header.Add("X-Broker-API-Originating-Identity", userIdentity)

	return req, nil
}

func doRequest(req *http.Request, httpc *http.Client, obj interface{}) (*http.Response, interface{}, *ServiceBrokerError) {
	res, err := httpc.Do(req)
	if err != nil {
		return nil, nil, NewServiceBrokerError(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return res, nil, NewServiceBrokerError(err)
	}

	if isErrorResponse(res) == true {
		sbError := &ServiceBrokerError{}
		json.Unmarshal(body, sbError)
		return res, nil, sbError
	}

	json.Unmarshal(body, obj)
	return res, obj, nil
}
