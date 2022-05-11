package controlplane

import (
	"fmt"
	"strings"

	"github.com/keptn/go-utils/pkg/api/models"
	api "github.com/keptn/go-utils/pkg/api/utils"
	keptnv2 "github.com/keptn/go-utils/pkg/lib/v0_2_0"
	"github.com/sirupsen/logrus"
)

type LogForwarder interface {
	Forward(keptnEvent models.KeptnContextExtendedCE, integrationID string) error
}

type LogForwardingHandler struct {
	logApi api.LogsV1Interface
}

func NewLogForwarder(logApi api.LogsV1Interface) *LogForwardingHandler {
	return &LogForwardingHandler{
		logApi: logApi,
	}
}

func (l LogForwardingHandler) Forward(keptnEvent models.KeptnContextExtendedCE, integrationID string) error {
	if strings.HasSuffix(*keptnEvent.Type, ".finished") {
		eventData := &keptnv2.EventData{}
		if err := keptnv2.EventDataAs(keptnEvent, eventData); err != nil {
			return fmt.Errorf("could not decode Keptn event data: %w", err)
		}

		taskName, _, _ := keptnv2.ParseTaskEventType(*keptnEvent.Type)

		if eventData.Status == keptnv2.StatusErrored {
			logrus.Info("Received '.finished' event with status 'errored'. Forwarding log message to log ingestion API")
			l.logApi.Log([]models.LogEntry{{
				IntegrationID: integrationID,
				Message:       eventData.Message,
				KeptnContext:  keptnEvent.Shkeptncontext,
				Task:          taskName,
				TriggeredID:   keptnEvent.Triggeredid,
			}})
		}
		return nil
	} else if *keptnEvent.Type == keptnv2.ErrorLogEventName {
		logrus.Info("Received 'log.error' event. Forwarding log message to log ingestion API")

		eventData := &keptnv2.ErrorLogEvent{}
		if err := keptnv2.EventDataAs(keptnEvent, eventData); err != nil {
			return fmt.Errorf("unable decode Keptn event data: %w", err)
		}

		integrationID := integrationID
		if eventData.IntegrationID != "" {
			// overwrite default integrationID if it has been set in the event
			integrationID = eventData.IntegrationID
		}
		l.logApi.Log([]models.LogEntry{{
			IntegrationID: integrationID,
			Message:       eventData.Message,
			KeptnContext:  keptnEvent.Shkeptncontext,
			Task:          eventData.Task,
			TriggeredID:   keptnEvent.Triggeredid,
		}})
	}
	return nil
}
