package consulnotifications

import (
	log "github.com/Sirupsen/logrus"
	"github.com/dpires/consul-leader-election"
	"github.com/hashicorp/consul/api"
	"time"
)

type HealthCheckMonitor struct {
	StopMonitorChannel chan bool
	Candidate          *election.LeaderElection
	Client             election.ConsulInterface
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
				options := &api.QueryOptions{WaitTime: time.Duration(monitor.Candidate.WatchWaitTime), WaitIndex: 0}
				healthchecks, err := monitor.Client.GetHealthChecks("any", options)

				if err != nil {
					log.Error(err)
				}

				for _, check := range healthchecks {
					switch check.Status {
					case "warning", "critical":
						log.Errorf("%s %s", check.Name, check.Status, check.Notes)
                                                // notification := NewNotification(check.Name, check.Status, check.Notes)
					}
				}
			} else {
				log.Info("Not Leader, returning")
			}
			time.Sleep(time.Duration(monitor.Candidate.WatchWaitTime) * time.Second)
		}
	}
}

func StartMonitor(channel chan bool, candidate *election.LeaderElection, client election.ConsulInterface) *HealthCheckMonitor {
	mon := &HealthCheckMonitor{
		StopMonitorChannel: channel,
		Candidate:          candidate,
		Client:             client,
	}
	go mon.StartMonitor()
	return mon
}
