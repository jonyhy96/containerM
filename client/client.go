package client

import (
	"log"

	dc "github.com/docker/docker/client"
)

var client *dc.Client

// GetCli get cli
func GetCli() *dc.Client {
	if client == nil {
		cli, err := dc.NewClientWithOpts(dc.FromEnv)
		if err != nil {
			log.Printf("GetCli err:%+v\n", err)
			return nil
		}
		client = cli
	}
	return client
}
