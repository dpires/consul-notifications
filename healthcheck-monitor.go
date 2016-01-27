package consulnotifications

import (
	"encoding/json"
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

func (monitor *HealthCheckMonitor) IsPassedTimeWindow(notification *Notification, window string) bool {
	checkTime := notification.Created
	duration, _ := time.ParseDuration(window)
	if time.Since(checkTime) >= duration {
		return true
	}
	return false
}

func (monitor *HealthCheckMonitor) StoreNotification(notification *Notification) error {
	value, _ := json.Marshal(notification)
	kv := &api.KVPair{Key: notification.Id, Value: value}
	err := monitor.Client.PutKey(kv)
	return err
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
							not := Notification{}
							json.Unmarshal(aquired.Value, &not)
							if monitor.IsPassedTimeWindow(&not, "5s") {
								if not.Sent == false {
									log.Info("SENDING ALERT")
								} else {
									log.Infof("alert status = %b", not.Sent)
								}
								not.Sent = true
								err = monitor.StoreNotification(&not)
								if err != nil {
									log.Error(err)
								}
							}

						} else {
							value := NewNotification(key, "test", "test")
							err = monitor.StoreNotification(value)
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
