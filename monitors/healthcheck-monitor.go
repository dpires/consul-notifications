package monitors

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dpires/consul-leader-election"
	"time"
)

type HealthCheckMonitor struct {
	StopMonitorChannel chan bool
	Candidate          *election.LeaderElection
        Client *election.ConsulInterface
}

func (monitor *HealthCheckMonitor) StartMonitor() {
	stop := false
	for !stop {
		select {
		case <-monitor.StopMonitorChannel:
			stop = true
			log.Info("Stopping HealthCheck monitor")
		default:
			if monitor.Candidate.IsLeader() {
				log.Info("Checking healthchecks")
			} else {
				log.Info("Not Leader, returning")
			}
			time.Sleep(time.Duration(monitor.Candidate.WatchWaitTime) * time.Second)
		}
	}
}

func StartMonitor(channel chan bool, candidate *election.LeaderElection, client *election.ConsulInterface) *HealthCheckMonitor {
	mon := &HealthCheckMonitor{
		StopMonitorChannel: channel,
		Candidate:          candidate,
                Client: client,
	}
	go mon.StartMonitor()
	return mon
}
