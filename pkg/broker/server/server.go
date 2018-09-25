/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	"net/http"
	"time"

	brokerapi "github.com/aerogear/managed-services-broker/pkg/broker"
	"github.com/aerogear/managed-services-broker/pkg/broker/controller"
	"github.com/aerogear/managed-services-broker/pkg/broker/server/util"

	"github.com/gorilla/mux"
	glog "github.com/sirupsen/logrus"
)


type server struct {
	controller controller.Controller
}

// CreateHandler creates Broker HTTP handler based on an implementation
// of a controller.Controller interface.
func createHandler(c controller.Controller) http.Handler {
	s := server{
		controller: c,
	}

	var router = mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.catalog).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/last_operation", s.getServiceInstanceLastOperation).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", s.createServiceInstance).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}", s.removeServiceInstance).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", s.bind).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", s.unBind).Methods("DELETE")
	router.Use(loggingMiddleware)

	return router
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		glog.Infof("%s to %s", r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// Run creates the HTTP handler based on an implementation of a
// controller.Controller interface, and begins to listen on the specified address.
func Run(ctx context.Context, addr string, c controller.Controller) error {
	listenAndServe := func(srv *http.Server) error {
		return srv.ListenAndServe()
	}
	return run(ctx, addr, listenAndServe, c)
}

// RunTLS creates the HTTPS handler based on an implementation of a
// controller.Controller interface, and begins to listen on the specified address.
func RunTLS(ctx context.Context, addr string, cert string, key string, c controller.Controller) error {
	var tlsCert tls.Certificate
	var err error
	tlsCert, err = tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return err
	}
	listenAndServe := func(srv *http.Server) error {
		srv.TLSConfig = new(tls.Config)
		srv.TLSConfig.Certificates = []tls.Certificate{tlsCert}
		return srv.ListenAndServeTLS("", "")
	}
	return run(ctx, addr, listenAndServe, c)
}

func run(ctx context.Context, addr string, listenAndServe func(srv *http.Server) error, c controller.Controller) error {
	glog.Infof("Starting server on %s", addr)
	srv := &http.Server{
		Addr:    addr,
		Handler: createHandler(c),
	}
	go func() {
		<-ctx.Done()
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if srv.Shutdown(c) != nil {
			srv.Close()
		}
	}()
	return listenAndServe(srv)
}

func (s *server) catalog(w http.ResponseWriter, r *http.Request) {
	glog.Infof("Get service broker catalog...")

	if result, err := s.controller.Catalog(); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
	}
}

func (s *server) getServiceInstanceLastOperation(w http.ResponseWriter, r *http.Request) {
	instanceID := mux.Vars(r)["instance_id"]
	q := r.URL.Query()
	serviceID := q.Get("service_id")
	planID := q.Get("plan_id")
	operation := q.Get("operation")
	glog.Infof("Get service slice... %s", instanceID)

	if result, err := s.controller.GetServiceInstanceLastOperation(instanceID, serviceID, planID, operation); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
	}
}

func (s *server) createServiceInstance(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["instance_id"]
	var req brokerapi.CreateServiceInstanceRequest
	if err := util.BodyToObject(r, &req); err != nil {
		glog.Errorf("error unmarshalling: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}
	if err := util.GetOriginatingUserInfo(r, &req.OriginatingUserInfo); err != nil {
		glog.Errorf("error retrieving originating user info: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
	}

	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}


	q := r.URL.Query()
	async := q.Get("accepts_incomplete") == "true"
	if async != true {
		util.WriteResponse(w, http.StatusUnprocessableEntity, brokerapi.NewUnprocessableEntityError())
		return
	}

	serviceID := req.Parameters["service_id"]
	if serviceID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid service_id"))
		return
	}

	planID := req.Parameters["plan_id"]
	if planID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid plan_id"))
		return
	}

	result, err := s.controller.CreateServiceInstance(id, &req)
	if err != nil {
		// Should handle:
		// if the Service Instance already exists error status code 200
		// if a Service Instance with the same id already exists but with different attributes error status code 409
		util.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if async == true {
		// Return an identifier representing the operation.
		util.WriteResponse(w, http.StatusAccepted, result)
	} else {
		util.WriteResponse(w, http.StatusCreated, result)
	}
}

func (s *server) removeServiceInstance(w http.ResponseWriter, r *http.Request) {
	glog.Infof("r: %+v", r)

	instanceID := mux.Vars(r)["instance_id"]

	q := r.URL.Query()
	serviceID := q.Get("service_id")
	planID := q.Get("plan_id")
	async := q.Get("accepts_incomplete") == "true"

	if instanceID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid instance_uuid"))
		return
	}

	if serviceID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid service_id"))
		return
	}

	if planID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid plan_id"))
		return
	}

	if planID == "" {
		util.WriteErrorResponse(w, http.StatusBadRequest, errors.NewBadRequest("invalid plan_id"))
		return
	}

	if async != true {
		util.WriteResponse(w, http.StatusUnprocessableEntity, brokerapi.NewUnprocessableEntityError())
		return
	}

	result, err := s.controller.RemoveServiceInstance(instanceID, serviceID, planID, async);
	if err != nil {
		// Should handle:
		// if the Service Instance does not exist status code 410
		util.WriteErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if async == true {
		// Return an identifier representing the operation.
		util.WriteResponse(w, http.StatusAccepted, result)
	} else {
		util.WriteResponse(w, http.StatusOK, result)
	}
}

func (s *server) bind(w http.ResponseWriter, r *http.Request) {
	bindingID := mux.Vars(r)["binding_id"]
	instanceID := mux.Vars(r)["instance_id"]

	glog.Infof("Bind binding_id=%s, instance_id=%s\n", bindingID, instanceID)

	var req brokerapi.BindingRequest

	if err := util.BodyToObject(r, &req); err != nil {
		glog.Errorf("Failed to unmarshall request: %v", err)
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	// TODO: Check if parameters are required, if not, this thing below is ok to leave in,
	// if they are ,they should be checked. Because if no parameters are passed in, this will
	// fail because req.Parameters is nil.
	if req.Parameters == nil {
		req.Parameters = make(map[string]interface{})
	}

	// Pass in the instanceId to the template.
	req.Parameters["instanceId"] = instanceID

	if result, err := s.controller.Bind(instanceID, bindingID, &req); err == nil {
		util.WriteResponse(w, http.StatusOK, result)
	} else {
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
	}
}

func (s *server) unBind(w http.ResponseWriter, r *http.Request) {
	instanceID := mux.Vars(r)["instance_id"]
	bindingID := mux.Vars(r)["binding_id"]
	q := r.URL.Query()
	serviceID := q.Get("service_id")
	planID := q.Get("plan_id")
	glog.Infof("UnBind: Service instance guid: %s:%s", bindingID, instanceID)

	if err := s.controller.UnBind(instanceID, bindingID, serviceID, planID); err == nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "{}") //id)
	} else {
		util.WriteErrorResponse(w, http.StatusBadRequest, err)
	}
}
