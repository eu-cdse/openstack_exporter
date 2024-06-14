package internal

import (
	"context"
	"github.com/go-kit/log/level"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
)

func authenticateOpenStack() (*gophercloud.ProviderClient, error) {
	ctx := context.Background()
	// Load OpenStack client options from the environment
	level.Debug(logger).Log("message", "Parsing environment variables")
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		level.Error(logger).Log("message", "Failed to parse environment variables", "err", err)
	}

	// Authenticate with OpenStack
	level.Debug(logger).Log("message", "Authenticating to OpenStack API")
	providerClient, err := openstack.AuthenticatedClient(ctx, opts)
	if err != nil {
		level.Error(logger).Log("message", "Failed to authenticate to OpenStack API", "err", err)
	}
	return providerClient, nil
}
