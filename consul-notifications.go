package main

import (
	"os"
	"os/signal"
	"github.com/dpires/consul-leader-election"
	"github.com/dpires/consul-leader-election/client"
	"github.com/dpires/consul-notifications/monitors"
	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

func main() {
	config := api.DefaultConfig()
	consulClient, _ := api.NewClient(config)

	leaderElection := &election.LeaderElection{
		StopElection:  make(chan bool),
		LeaderKey:     "service/consul-notifications/leader",
		WatchWaitTime: 3,
		Client: &client.ConsulClient{
			Client: consulClient,
		},
	}

	go leaderElection.ElectLeader()

	mon := monitors.StartMonitor(make(chan bool), leaderElection)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)
	select {
	case <-signalChannel:
		mon.StopMonitorChannel <- true
		leaderElection.CancelElection()
		log.Info("Exiting Consul Notifications")
	}
}
