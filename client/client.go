package client

import (
	"io"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/anthontaylor/skipper"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

// New returns a service that's load-balanced over instances of skipper found
// in the provided Consul server. The mechanism of looking up skipper
// instances in Consul is hard-coded into the client.
func New(consulAddr string, logger log.Logger) (skipper.Service, error) {
	apiclient, err := consulapi.NewClient(&consulapi.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}

	// As the implementer of skipper, we declare and enforce these
	// parameters for all of the skipper consumers.
	var (
		consulService = "skipper"
		consulTags    = []string{"prod"}
		passingOnly   = true
		retryMax      = 3
		retryTimeout  = 500 * time.Millisecond
	)

	var (
		sdclient  = consul.NewClient(apiclient)
		instancer = consul.NewInstancer(sdclient, logger, consulService, consulTags, passingOnly)
		endpoints skipper.Endpoints
	)
	{
		factory := factoryFor(skipper.MakePostProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostProfileEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakeGetProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetProfileEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakePutProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PutProfileEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakePatchProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PatchProfileEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakeDeleteProfileEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteProfileEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakeGetAddressesEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressesEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakeGetAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.GetAddressEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakePostAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.PostAddressEndpoint = retry
	}
	{
		factory := factoryFor(skipper.MakeDeleteAddressEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.DeleteAddressEndpoint = retry
	}

	return endpoints, nil
}

func factoryFor(makeEndpoint func(skipper.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		service, err := skipper.MakeClientEndpoints(instance)
		if err != nil {
			return nil, nil, err
		}
		return makeEndpoint(service), nil, nil
	}
}
