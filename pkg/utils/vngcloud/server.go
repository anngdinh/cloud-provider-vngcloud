package vngcloud

import (
	"github.com/vngcloud/vngcloud-go-sdk/client"
	"github.com/vngcloud/vngcloud-go-sdk/vngcloud/objects"
	"github.com/vngcloud/vngcloud-go-sdk/vngcloud/services/compute/v2/server"
	"k8s.io/klog/v2"
)

func GetServer(client *client.ServiceClient, projectID string, id string) (*objects.Server, error) {
	klog.V(5).Infoln("[API] GetServer")
	opt := server.NewGetOpts(projectID, id)
	resp, err := server.Get(client, opt)
	if err != nil {
		return nil, err.Error
	}
	return resp, nil
}

func ListProviderID(client *client.ServiceClient, projectID string, providerIDs []string) ([]*objects.Server, error) {
	var servers []*objects.Server
	for _, providerID := range providerIDs {
		vServer, err := GetServer(client, projectID, providerID)
		if err != nil {
			return nil, err
		} else {
			servers = append(servers, vServer)
		}
	}
	return servers, nil
}
