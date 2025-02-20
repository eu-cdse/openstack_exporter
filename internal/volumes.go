package internal

import (
	"context"
	"encoding/json"
	"os"

	"github.com/go-kit/log/level"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	volumeLimits "github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v2/limits"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v2/volumes"
)

type Volume struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type AbsoluteVolumeLimits struct {
	MaxTotalVolumes         int `json:"maxTotalVolumes"`
	MaxTotalVolumeGigabytes int `json:"maxTotalVolumeGigabytes"`
	TotalVolumesUsed        int `json:"totalVolumesUsed"`
	TotalGigabytesUsed      int `json:"totalGigabytesUsed"`
}

func getVolumeLimits(providerClient *gophercloud.ProviderClient) (*volumeLimits.Limits, error) {
	var blockStorageClient *gophercloud.ServiceClient
	var err error
	if os.Getenv("OS_BLOCKSTORAGE_V") == "2" {
		blockStorageClient, err = openstack.NewBlockStorageV2(providerClient, gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		})
	} else {
		blockStorageClient, err = openstack.NewBlockStorageV3(providerClient, gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		})
	}
	if err != nil {
		return nil, err
	}

	level.Debug(logger).Log("message", "Getting volume limits")

	volumeLimits, err := volumeLimits.Get(context.Background(), blockStorageClient).Extract()
	if err != nil {
		return nil, err
	}

	return volumeLimits, nil
}

func getAllVolumes(providerClient *gophercloud.ProviderClient) []volumes.Volume {
	blockStorageClient, err := openstack.NewBlockStorageV3(providerClient, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve volumes", "err", err)
	}

	listOpts := volumes.ListOpts{
		AllTenants: false,
	}

	level.Debug(logger).Log("message", "Getting all volumes")

	allPages, err := volumes.List(blockStorageClient, listOpts).AllPages(context.TODO())
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all volumes", "err", err)
	}

	allVolumes, err := volumes.ExtractVolumes(allPages)
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all volumes", "err", err)
	}

	return allVolumes
}

func countVolumePerStatus(volumeList []volumes.Volume) map[string]int {
	volumeListJSON, err := json.Marshal(volumeList)
	if err != nil {
		level.Error(logger).Log("message", "Error parsing volume list", "err", err)
	}

	var volumes []Volume
	if err := json.Unmarshal([]byte(volumeListJSON), &volumes); err != nil {
		level.Error(logger).Log("message", "Error parsing JSON", "err", err)
	}
	statusCount := make(map[string]int)

	for _, volume := range volumes {
		status := volume.Status
		statusCount[status]++
	}

	return statusCount
}
