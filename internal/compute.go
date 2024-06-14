package internal

import (
	"context"
	"encoding/json"
	"os"

	"github.com/go-kit/log/level"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	computeLimits "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/limits"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

type AbsoluteComputeLimits struct {
	MaxTotalCores      int `json:"maxTotalCores"`
	MaxTotalInstances  int `json:"maxTotalInstances"`
	MaxTotalRAMSize    int `json:"maxTotalRAMSize"`
	TotalCoresUsed     int `json:"totalCoresUsed"`
	TotalInstancesUsed int `json:"totalInstancesUsed"`
	TotalRAMUsed       int `json:"totalRAMUsed"`
}

type Server struct {
	ID     string                 `json:"ID"`
	Flavor map[string]interface{} `json:"Flavor"`
	Status string                 `json:"Status"`
}

func getAllServers(providerClient *gophercloud.ProviderClient) []servers.Server {
	// Create a Compute V2 service client
	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
	if err != nil {
		level.Error(logger).Log("message", "Failed to create compute client", "err", err)
	}
	listOpts := servers.ListOpts{
		AllTenants: false,
	}

	level.Debug(logger).Log("message", "Getting all servers")

	allPages, err := servers.List(computeClient, listOpts).AllPages(context.TODO())
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all servers", "err", err)
	}

	allServers, err := servers.ExtractServers(allPages)
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all servers", "err", err)
	}

	return allServers
}

func countInstancePerFlavor(serverList []servers.Server) map[string]int {
	serverListJSON, err := json.Marshal(serverList)
	if err != nil {
		level.Error(logger).Log("message", "Error parsing server list", "err", err)
	}

	var servers []Server
	if err := json.Unmarshal([]byte(serverListJSON), &servers); err != nil {
		level.Error(logger).Log("message", "Error parsing JSON", "err", err)
	}
	flavorCount := make(map[string]int)

	for _, server := range servers {
		flavorID := server.Flavor["id"].(string)
		flavorCount[flavorID]++
	}

	return flavorCount
}

func countInstancePerStatus(serverList []servers.Server) map[string]int {
	serverListJSON, err := json.Marshal(serverList)
	if err != nil {
		level.Error(logger).Log("message", "Error parsing server list", "err", err)
	}

	var servers []Server
	if err := json.Unmarshal([]byte(serverListJSON), &servers); err != nil {
		level.Error(logger).Log("message", "Error parsing JSON", "err", err)
	}
	statusCount := make(map[string]int)

	for _, server := range servers {
		status := server.Status
		statusCount[status]++
	}

	return statusCount
}
func getComputeLimits(providerClient *gophercloud.ProviderClient) *computeLimits.Limits {
	// Create a Compute V2 service client
	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
	if err != nil {
		level.Error(logger).Log("message", "Failed to create compute client", "err", err)
	}

	getOpts := computeLimits.GetOpts{}

	level.Debug(logger).Log("message", "Getting compute limits")
	computeLimits, err := computeLimits.Get(context.TODO(), computeClient, getOpts).Extract()
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve limits", "err", err)
	}

	return computeLimits
}
