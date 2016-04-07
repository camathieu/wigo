//package notify
//import (
//	"github.com/root-gg/wigo/wigo/runner"
//	"github.com/root-gg/wigo/wigo/global"
//)
//
//
//type NotificationBackend interface {
//	Send(notif Notification)
//}
//
//func RegisterBackends(){
//
//}
//
//func notify() {
//
//}
//
//// Notifications
//type INotification interface {
//	Json() string
//	Message() string
//	Summary() string
//}
//
//type Notification struct {
//	Notification
//}
//
//type NotificationProbe struct {
//	*Notification
//}
//
//type NotificationWigo struct {
//	*Notification
//}
//
//
//func NewNotification() (n *Notification) {
//	n = new(Notification)
//	return
//}
//
//type NotificationProbe struct {
//	*Notification
//	Host *global.Wigo
//	oldProbe *runner.ProbeResult
//	newProbe *runner.ProbeResult
//}
//
//func NewNotificationProbe(wigo *global.Wigo, oldProbe *runner.ProbeResult, newProbe *runner.ProbeResult) (np *NotificationProbe) {
//	np = new(NotificationProbe)
//	np.Notification = NewNotification()
//	np.Host = wigo
//	np.OldProbe = oldProbe
//	np.NewProbe = newProbe
//
//	if oldProbe == nil && newProbe != nil {
//		np.Hostname = newProbe.GetHost().GetParentWigo().Hostname
//		this.Message = fmt.Sprintf("New probe %s with status %d detected on host %s", newProbe.Name, newProbe.Status, this.Hostname)
//
//		this.Summary += fmt.Sprintf("A new probe %s has been detected on host %s : \n\n", newProbe.Name, this.Hostname)
//		this.Summary += fmt.Sprintf("\t%s\n", newProbe.Message)
//
//	} else if oldProbe != nil && newProbe == nil {
//		this.Hostname = oldProbe.GetHost().GetParentWigo().Hostname
//		this.Message = fmt.Sprintf("Probe %s on host %s does not exist anymore. Last status was %d", oldProbe.Name, this.Hostname, oldProbe.Status)
//
//		this.Summary += fmt.Sprintf("Probe %s has been deleted on host %s : \n\n", oldProbe.Name, this.Hostname)
//		this.Summary += fmt.Sprintf("Last message was : \n\n%s\n", oldProbe.Message)
//
//	} else if oldProbe != nil && newProbe != nil {
//		if newProbe.Status != oldProbe.Status {
//			this.Hostname = newProbe.GetHost().GetParentWigo().Hostname
//
//			this.Message = fmt.Sprintf("Probe %s status changed from %d to %d on host %s", newProbe.Name, oldProbe.Status, newProbe.Status, this.Hostname)
//
//			this.Summary += fmt.Sprintf("Probe %s on host %s : \n\n", oldProbe.Name, this.Hostname)
//			this.Summary += fmt.Sprintf("\tOld Status : %d\n", oldProbe.Status)
//			this.Summary += fmt.Sprintf("\tNew Status : %d\n\n", newProbe.Status)
//			this.Summary += fmt.Sprintf("Message :\n\n\t%s\n\n", newProbe.Message)
//
//			// List parent host probes in error
//			this.HostProbesInError = newProbe.parentHost.GetErrorsProbesList()
//
//			// Add Log
//			LocalWigo.AddLog(newProbe, INFO, fmt.Sprintf("Probe %s switched from %d to %d : %s", newProbe.Name, oldProbe.Status, newProbe.Status, newProbe.Message))
//		}
//	}
//
//	return
//}
//
//type NotificationWigo struct {
//	Notification
//}
//
//func NewNotificationWigo() (nw *NotificationWigo) {
//	nw = new(NotificationWigo)
//	nw.Notification = NewNotification()
//	return
//}
//
//
//func NewProbeNotification(oldProbe *runner.ProbeResult, newProbe *runner.ProbeResult) (this *Notification) {
//	this = new(NotificationProbe)
//	this.Notification = NewNotification()
//	this.Type = "Probe"
//	this.OldProbe = oldProbe
//	this.NewProbe = newProbe
//
//	if oldProbe == nil && newProbe != nil {
//		this.Hostname = newProbe.GetHost().GetParentWigo().Hostname
//		this.Message = fmt.Sprintf("New probe %s with status %d detected on host %s", newProbe.Name, newProbe.Status, this.Hostname)
//
//		this.Summary += fmt.Sprintf("A new probe %s has been detected on host %s : \n\n", newProbe.Name, this.Hostname)
//		this.Summary += fmt.Sprintf("\t%s\n", newProbe.Message)
//
//	} else if oldProbe != nil && newProbe == nil {
//		this.Hostname = oldProbe.GetHost().GetParentWigo().Hostname
//		this.Message = fmt.Sprintf("Probe %s on host %s does not exist anymore. Last status was %d", oldProbe.Name, this.Hostname, oldProbe.Status)
//
//		this.Summary += fmt.Sprintf("Probe %s has been deleted on host %s : \n\n", oldProbe.Name, this.Hostname)
//		this.Summary += fmt.Sprintf("Last message was : \n\n%s\n", oldProbe.Message)
//
//	} else if oldProbe != nil && newProbe != nil {
//		if newProbe.Status != oldProbe.Status {
//			this.Hostname = newProbe.GetHost().GetParentWigo().Hostname
//
//			this.Message = fmt.Sprintf("Probe %s status changed from %d to %d on host %s", newProbe.Name, oldProbe.Status, newProbe.Status, this.Hostname)
//
//			this.Summary += fmt.Sprintf("Probe %s on host %s : \n\n", oldProbe.Name, this.Hostname)
//			this.Summary += fmt.Sprintf("\tOld Status : %d\n", oldProbe.Status)
//			this.Summary += fmt.Sprintf("\tNew Status : %d\n\n", newProbe.Status)
//			this.Summary += fmt.Sprintf("Message :\n\n\t%s\n\n", newProbe.Message)
//
//			// List parent host probes in error
//			this.HostProbesInError = newProbe.parentHost.GetErrorsProbesList()
//
//			// Add Log
//			LocalWigo.AddLog(newProbe, INFO, fmt.Sprintf("Probe %s switched from %d to %d : %s", newProbe.Name, oldProbe.Status, newProbe.Status, newProbe.Message))
//		}
//	}
//
//	// Log
//	log.Printf("New Probe Notification : %s", this.Message)
//
//	// Send ?
//	if GetLocalWigo().GetConfig().Notifications.OnProbeChange {
//		weSend := false
//
//		if oldProbe != nil && newProbe != nil {
//			if newProbe.Status < oldProbe.Status && oldProbe.Status >= GetLocalWigo().GetConfig().Notifications.MinLevelToSend {
//				// It's an UP
//				weSend = true
//			} else if newProbe.Status >= GetLocalWigo().GetConfig().Notifications.MinLevelToSend {
//				// It's a DOWN, check if new status is > to MinLevelToSend
//				weSend = true
//			}
//		}
//
//		if weSend {
//			Channels.ChanCallbacks <- this
//		}
//	}
//
//	return
//}
