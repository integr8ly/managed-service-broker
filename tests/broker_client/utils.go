package broker_client

import (
	"net/http"
	"time"
)

func NewServiceBrokerError(err error) *ServiceBrokerError {
	return &ServiceBrokerError{
		Description: err.Error(),
	}
}

func NewServiceBrokerClient(cCfg *BrokerClientClientConfig) *ServiceBrokerClient {
	sbc := &ServiceBrokerClient{
		HttpClient: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: cCfg.TlsCfg,
			},
		},
		BrokerURL:    cCfg.BrokerURL,
		Token:        cCfg.Token,
		UserIdentity: cCfg.UserIdentity,
	}

	return sbc
}

func isErrorResponse(res *http.Response) bool {
	return res.StatusCode < http.StatusOK || res.StatusCode > http.StatusNoContent
}
