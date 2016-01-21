package consulnotifications

import (
	"fmt"
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
						log.Errorf("%s %s %s", check.Name, check.Status, check.Notes)
						if check.ServiceID == "" {
							check.ServiceID = "no-service"
						}
						key := fmt.Sprintf("consul-notifications/health-checks/%s/%s/%s", check.Node, check.ServiceID, check.CheckID)
						log.Info(key)
						aquired, err := monitor.Client.GetKey(key)
						if err != nil {
							log.Error(err)
						}

						if aquired != nil {
							log.Infof("Aquired Key %s:", key)
							checkTime := string(aquired.Value)
							log.Infof("stored time=%s", checkTime)
							timeVal, _ := time.Parse(time.UnixDate, checkTime)

							duration, _ := time.ParseDuration("10s")
							// check elapsed time, notify if over
							if time.Since(timeVal) >= duration {
								// notification := NewNotification(check.Name, check.Status, check.Notes)
								log.Info("SENDING ALERT")
							}

						} else {
							log.Infof("Key not aquired, aquiring...")
							t := time.Now()
							ee := t.Format(time.UnixDate)
							value := []byte(ee)
							log.Infof("Storing time = %s", ee)
							kv := &api.KVPair{Key: key, Value: value}
							err = monitor.Client.PutKey(kv)
							if err != nil {
								log.Error(err)
							}
						}
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
