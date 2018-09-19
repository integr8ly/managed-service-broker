package broker_client

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewServiceBrokerError(err error) *ServiceBrokerError {
	return &ServiceBrokerError{
		Description: err.Error(),
	}
}

func NewServiceBrokerClient(brokerURL, token, userIdentity string, insecureSkipVerify bool) *ServiceBrokerClient{
	sbc := &ServiceBrokerClient {
		HttpClient: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecureSkipVerify,
				},
			},
		},
		BrokerURL: brokerURL,
		Token:     token,
		UserIdentity: userIdentity,
	}

	return sbc
}

func isErrorResponse(res *http.Response) bool {
	return res.StatusCode < http.StatusOK || res.StatusCode > http.StatusNoContent

}