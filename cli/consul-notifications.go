package main

import (
	"github.com/dpires/consul-leader-election"
	"github.com/dpires/consul-leader-election/client"
	"github.com/hashicorp/consul/api"
	cn "github.com/dpires/consul-notifications"
)

func main() {
	config := api.DefaultConfig()
	consulClient, _ := api.NewClient(config)

        consulInterface := &client.ConsulClient{
                Client: consulClient,
        }

	leaderElection := &election.LeaderElection{
		StopElection:  make(chan bool),
		LeaderKey:     "service/consul-notifications/leader",
		WatchWaitTime: 3,
                Client: consulInterface,
	}

        app := &cn.ConsulNotifications{
            ConsulClient: consulInterface,
            Leader: leaderElection,
        } 

        app.Start()
}
