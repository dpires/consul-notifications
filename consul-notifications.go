package consulnotifications

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dpires/consul-leader-election"
	"os"
	"os/signal"
)

type ConsulNotifications struct {
	ConsulClient election.ConsulInterface
	Leader       *election.LeaderElection
}

func (cn *ConsulNotifications) Start() {

	go cn.Leader.ElectLeader()

	mon := StartMonitor(make(chan bool), cn.Leader, &cn.ConsulClient)

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, os.Kill)
	select {
	case <-signalChannel:
		mon.StopMonitorChannel <- true
		cn.Leader.CancelElection()
		log.Info("Exiting Consul Notifications")
	}
}
