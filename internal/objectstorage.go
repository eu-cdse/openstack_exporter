package internal

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/go-kit/log/level"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1/containers"
	gophertelekomcloud "github.com/opentelekomcloud/gophertelekomcloud"
	otc "github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/obs"
)

type Container struct {
	Bytes int64  `json:"bytes"`
	Name  string `json:"name"`
}

func newOBSClient() (*obs.ObsClient, error) {
	OSEnv := otc.NewEnv("OS_")
	providerClient, err := OSEnv.AuthenticatedClient()

	if err != nil {
		return nil, err
	}

	client, err := otc.NewOBSService(providerClient, gophertelekomcloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}
	//opts := cc.AKSKAuthOptions
	aksk := gophertelekomcloud.AKSKAuthOptions{
		IdentityEndpoint: os.Getenv("OS_AUTH_URL"),
		ProjectId:        os.Getenv("OS_PROJECT_ID"),
		ProjectName:      os.Getenv("OS_TENANT_NAME"),
		DomainID:         os.Getenv("OS_DOMAIN_ID"),
		Domain:           os.Getenv("OS_DOMAIN_NAME"),
		AccessKey:        os.Getenv("OS_ACCESS_KEY"),
		SecretKey:        os.Getenv("OS_SECRET_KEY"),
	}
	return obs.New(
		aksk.AccessKey, aksk.SecretKey, client.Endpoint,
		obs.WithSecurityToken(aksk.SecurityToken), obs.WithSignature(obs.SignatureObs),
	)
}

func getContainerList(providerClient *gophercloud.ProviderClient) []Container {
	// Create a ObjectStorage V1 service client

	var objectStorageClient *gophercloud.ServiceClient
	var err error
	if strings.Contains(os.Getenv("OS_AUTH_URL"), "otc") {
		level.Debug(logger).Log("message", "Setting up OBS client")

		obsClient, err := newOBSClient()

		if err != nil {
			level.Error(logger).Log("message", "Failed to setup OBS client", "err", err)
		}

		level.Debug(logger).Log("message", "Getting all containers")
		containerList, _ := obsClient.ListBuckets(nil)

		var containers []Container
		for _, bucket := range containerList.Buckets {
			bucketStorage, err := obsClient.GetBucketStorageInfo(bucket.Name)
			if err != nil {
				level.Error(logger).Log("message", "Error parsing container list", "err", err)
			}
			containers = append(containers, Container{Bytes: bucketStorage.Size, Name: bucket.Name})
		}
		return containers
	} else {
		objectStorageClient, err = openstack.NewObjectStorageV1(providerClient, gophercloud.EndpointOpts{
			Region: os.Getenv("OS_REGION_NAME"),
		})
		if err != nil {
			level.Error(logger).Log("message", "Failed to create objectstorage client", "err", err)
		}
	}

	listOpts := containers.ListOpts{}

	level.Debug(logger).Log("message", "Getting all containers")

	allPages, err := containers.List(objectStorageClient, listOpts).AllPages(context.TODO())
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all containers", "err", err)
	}

	containerList, err := containers.ExtractInfo(allPages)
	if err != nil {
		level.Error(logger).Log("message", "Failed to retrieve all containers", "err", err)
	}

	containerListJSON, err := json.Marshal(containerList)

	if err != nil {
		level.Error(logger).Log("message", "Error parsing container list", "err", err)
	}

	var containers []Container
	if err := json.Unmarshal([]byte(containerListJSON), &containers); err != nil {
		level.Error(logger).Log("message", "Error parsing JSON", "err", err)
		return nil
	}

	return containers
}
