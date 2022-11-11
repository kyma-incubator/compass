package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Action string

const (
	Insert Action = "INSERT"
	Delete Action = "DELETE"
	Update Action = "UPDATE"
)

type Table string

const (
	Applications         Table = "applications"
	Runtimes             Table = "runtimes"
	Labels               Table = "labels"
	Formations           Table = "formations"
	FormationAssignments Table = "formation_assignments"
)

type Notification struct {
	Table  Table
	Action Action
	Data   json.RawMessage
}

func (n *Notification) Validate() error {
	if n.Table == "" {
		return errors.New("missing table name")
	}
	if n.Action == "" {
		return errors.New("missing action")
	}

	if len(n.Data) == 0 {
		return errors.New("empty data")
	}

	return nil
}

type HandlerKey struct {
	NotificationChannel string
	ResourceType        resource.Type
}

func NewNotificationProcessor(cfg persistence.DatabaseConfig, handlers map[HandlerKey]NotificationHandler) *NotificationProcessor {
	return &NotificationProcessor{
		Handlers:      handlers,
		StorageConfig: cfg,
	}
}

type NotificationProcessor struct {
	listener      NotificationListener
	Handlers      map[HandlerKey]NotificationHandler
	StorageConfig persistence.DatabaseConfig
}

func (np *NotificationProcessor) Run(ctx context.Context) error {
	if err := np.connect(ctx); err != nil {
		return errors.Errorf("failed to connect notification processor: %s", err)
	}

	np.processLoop(ctx)

	return nil
}

func (np *NotificationProcessor) connect(ctx context.Context) error {
	reporter := func(event pq.ListenerEventType, err error) {
		switch event {
		case pq.ListenerEventConnected:
			log.C(ctx).Info("storage ListenerEventConnected")
		case pq.ListenerEventReconnected:
			log.C(ctx).Info("storage ListenerEventReconnected")
		case pq.ListenerEventDisconnected:
			log.C(ctx).Error("storage ListenerEventDisconnected.")
		case pq.ListenerEventConnectionAttemptFailed:
			log.C(ctx).Error("storage ListenerEventConnectionAttemptFailed")
		}
	}

	connString := np.StorageConfig.GetConnString()
	listener := pq.NewListener(connString, time.Second*5, time.Minute*10, reporter)
	if err := listener.Listen("events"); err != nil {
		return errors.Errorf("failed to listen on channel %s: %s", "admiral", err)
	}

	np.listener = listener

	return nil
}

func (np *NotificationProcessor) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.D().Debug("Stopping notifications processor...")
			if err := np.listener.Close(); err != nil {
				log.D().Errorf("Closing notifications processor returned error : %s", err)
				return
			}
		case n := <-np.listener.NotificationChannel():
			log.C(ctx).Infof("Received data from channel [%s]", n.Channel)
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				log.C(ctx).WithError(err).Error("failed JSON processing data from channel %s", n.Channel)
				continue
			}

			log.C(ctx).Infof("Received data from channel %s: %s", n.Channel, string(prettyJSON.Bytes()))

			notification := &Notification{}
			if err := json.Unmarshal([]byte(n.Extra), notification); err != nil {
				log.C(ctx).WithError(err).Error("failed to unmarshal notification")
				continue
			}

			if err := notification.Validate(); err != nil {
				log.C(ctx).WithError(err).Error("invalid notification")
				continue
			}

			resourceType, err := tableNameToResourceType(notification.Table)
			if err != nil {
				log.C(ctx).WithError(err).Error("failed convertion")
				continue
			}

			currentKey := HandlerKey{
				NotificationChannel: n.Channel,
				ResourceType:        resourceType,
			}

			notificationHandler, found := np.Handlers[currentKey]
			if !found {
				log.C(ctx).Errorf("Could not find notification handler for key %v", currentKey)
				continue
			}

			switch notification.Action {
			case Insert:
				if err := notificationHandler.HandleCreate(ctx, notification.Data); err != nil {
					log.C(ctx).WithError(err).Error("error during notification handling")
					continue
				}
			case Update:
				if err := notificationHandler.HandleUpdate(ctx, notification.Data); err != nil {
					log.C(ctx).WithError(err).Error("error during notification handling")
					continue
				}
			case Delete:
				if err := notificationHandler.HandleDelete(ctx, notification.Data); err != nil {
					log.C(ctx).WithError(err).Error("error during notification handling")
					continue
				}
			}

			log.C(ctx).Info("Successfully processed notification %s", string(prettyJSON.Bytes()))
		case <-time.After(90 * time.Second):
			log.C(ctx).Warn("Received no events for 90 seconds, checking connection")
			go func() {
				if err := np.listener.Ping(); err != nil {
					log.C(ctx).WithError(err).Error("pinging listener failed")
				}
			}()
			log.C(ctx).Warn("Successfully checking watch connection")
		}
	}
}

func tableNameToResourceType(table Table) (resource.Type, error) {
	switch table {
	case Applications:
		return resource.Application, nil
	case Runtimes:
		return resource.Runtime, nil
	case Labels:
		return resource.Label, nil
	case Formations:
		return resource.Formations, nil
	case FormationAssignments:
		return resource.FormationAssignment, nil
	}

	return "", errors.Errorf("failed to convert %s to resource type", table)
}
