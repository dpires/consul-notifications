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

type ConsulNotifications struct {
    ConsulClient election.ConsulInterface
    Leader *election.LeaderElection
}

func (cn *ConsulNotifications) Start() {

	go cn.Leader.ElectLeader()

	mon := monitors.StartMonitor(make(chan bool), cn.Leader)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)
	select {
	case <-signalChannel:
		mon.StopMonitorChannel <- true
		cn.Leader.CancelElection()
		log.Info("Exiting Consul Notifications")
	}
}

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

        app := &ConsulNotifications{
            ConsulClient: consulInterface,
            Leader: leaderElection,
        } 

        app.Start()
}
