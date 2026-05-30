package graph

import (
	"time"

	"github.com/tuankhanhvo/pulseops/graph/model"
	"github.com/tuankhanhvo/pulseops/internal/analytics"
	"github.com/tuankhanhvo/pulseops/internal/incidents"
	"github.com/tuankhanhvo/pulseops/internal/streams"
)

func mapIncidentEvent(event streams.IncidentEvent) (*model.IncidentEvent, bool) {
	doc, ok := incidentDocFromPayload(event.Payload)
	if !ok {
		return nil, false
	}

	return &model.IncidentEvent{
		Type:       model.IncidentEventType(event.Type),
		Incident:   MapIncidentDoc(doc),
		OccurredAt: time.Now().UTC(),
	}, true
}

func MapIncidentDoc(doc *incidents.IncidentDoc) *model.Incident {
	if doc == nil {
		return nil
	}

	return &model.Incident{
		ID:             doc.ID.Hex(),
		Title:          doc.Title,
		Status:         model.IncidentStatus(doc.Status),
		Severity:       model.Severity(doc.Severity),
		TeamID:         doc.TeamID.Hex(),
		Fingerprint:    doc.Fingerprint,
		AlertCount:     doc.AlertCount,
		TriggeredAt:    doc.TriggeredAt,
		AcknowledgedAt: doc.AcknowledgedAt,
		ResolvedAt:     doc.ResolvedAt,
		Escalated:      doc.Escalated,
		EscalatedAt:    doc.EscalatedAt,
		StatusMessage:  doc.StatusMessage,
		Mttr:           doc.MTTR,
		Alerts:         []*model.Alert{},
	}
}

func MapIncidentDocs(docs []*incidents.IncidentDoc) []*model.Incident {
	models := make([]*model.Incident, 0, len(docs))
	for _, doc := range docs {
		models = append(models, MapIncidentDoc(doc))
	}

	return models
}

func MapAlertDoc(doc *incidents.AlertDoc) *model.Alert {
	if doc == nil {
		return nil
	}

	return &model.Alert{
		ID:          doc.ID.Hex(),
		IncidentID:  doc.IncidentID.Hex(),
		Source:      doc.Source,
		AlertName:   doc.AlertName,
		Severity:    model.Severity(doc.Severity),
		Environment: doc.Environment,
		Payload:     doc.Payload,
		Fingerprint: doc.Fingerprint,
		ReceivedAt:  doc.ReceivedAt,
	}
}

func MapAlertDocs(docs []*incidents.AlertDoc) []*model.Alert {
	alerts := make([]*model.Alert, 0, len(docs))
	for _, doc := range docs {
		alerts = append(alerts, MapAlertDoc(doc))
	}

	return alerts
}

func MapUserDoc(doc *incidents.UserDoc) *model.User {
	if doc == nil {
		return nil
	}

	var avatarURL *string
	if doc.AvatarURL != "" {
		avatarURL = &doc.AvatarURL
	}

	return &model.User{
		ID:            doc.ID.Hex(),
		Email:         doc.Email,
		Name:          doc.Name,
		AvatarURL:     avatarURL,
		TeamID:        doc.TeamID.Hex(),
		Role:          model.Role(doc.Role),
		GoogleSubject: doc.GoogleSubject,
		CreatedAt:     doc.CreatedAt,
	}
}

func MapUserDocs(docs []*incidents.UserDoc) []*model.User {
	users := make([]*model.User, 0, len(docs))
	for _, doc := range docs {
		users = append(users, MapUserDoc(doc))
	}

	return users
}

func MapTeamDoc(doc *incidents.TeamDoc, members []*incidents.UserDoc, schedule *model.OnCallSchedule) *model.Team {
	if doc == nil {
		return nil
	}

	apiKeyHint := doc.APIKeyHint
	return &model.Team{
		ID:             doc.ID.Hex(),
		Name:           doc.Name,
		Members:        MapUserDocs(members),
		OnCallSchedule: schedule,
		APIKeyHint:     &apiKeyHint,
	}
}

func MapRunbookDoc(doc *incidents.RunbookDoc) *model.Runbook {
	if doc == nil {
		return nil
	}

	return &model.Runbook{
		ID:        doc.ID.Hex(),
		TeamID:    doc.TeamID.Hex(),
		Title:     doc.Title,
		Content:   doc.Content,
		Tags:      doc.Tags,
		UpdatedAt: doc.UpdatedAt,
	}
}

func MapRunbookDocs(docs []*incidents.RunbookDoc) []*model.Runbook {
	runbooks := make([]*model.Runbook, 0, len(docs))
	for _, doc := range docs {
		runbooks = append(runbooks, MapRunbookDoc(doc))
	}

	return runbooks
}

func MapPostmortemDoc(doc *incidents.PostmortemDoc) *model.Postmortem {
	if doc == nil {
		return nil
	}

	return &model.Postmortem{
		ID:          doc.ID.Hex(),
		IncidentID:  doc.IncidentID.Hex(),
		AuthorID:    doc.AuthorID.Hex(),
		Summary:     doc.Summary,
		Timeline:    doc.Timeline,
		ActionItems: doc.ActionItems,
		CreatedAt:   doc.CreatedAt,
	}
}

func MapAnalyticsResult(result *analytics.AnalyticsResult) *model.Analytics {
	if result == nil {
		return &model.Analytics{}
	}

	byDay := make([]*model.DayStat, 0, len(result.ByDay))
	for _, stat := range result.ByDay {
		byDay = append(byDay, &model.DayStat{Date: stat.Date, Count: stat.Count})
	}

	bySeverity := make([]*model.SeverityStat, 0, len(result.BySeverity))
	for _, stat := range result.BySeverity {
		bySeverity = append(bySeverity, &model.SeverityStat{Severity: model.Severity(stat.Severity), Count: stat.Count})
	}

	return &model.Analytics{
		MttrSeconds: result.MTTRSeconds,
		MttaSeconds: result.MTTASeconds,
		TotalCount:  result.TotalCount,
		ByDay:       byDay,
		BySeverity:  bySeverity,
	}
}

func MapOnCallScheduleDoc(doc *incidents.OnCallScheduleDoc, current *incidents.UserDoc) *model.OnCallSchedule {
	if doc == nil {
		return nil
	}

	rotation := make([]*model.User, 0, len(doc.Rotation))
	for _, userID := range doc.Rotation {
		rotation = append(rotation, &model.User{ID: userID.Hex(), TeamID: doc.TeamID.Hex()})
	}

	overrides := make([]*model.ScheduleOverride, 0, len(doc.Overrides))
	for _, override := range doc.Overrides {
		overrides = append(overrides, &model.ScheduleOverride{
			ID:       override.ID.Hex(),
			User:     &model.User{ID: override.UserID.Hex(), TeamID: doc.TeamID.Hex()},
			StartsAt: override.StartsAt,
			EndsAt:   override.EndsAt,
			Reason:   override.Reason,
		})
	}

	return &model.OnCallSchedule{
		ID:            doc.ID.Hex(),
		TeamID:        doc.TeamID.Hex(),
		Rotation:      rotation,
		IntervalDays:  doc.IntervalDays,
		CycleStart:    doc.CycleStart,
		CurrentOnCall: MapUserDoc(current),
		Overrides:     overrides,
	}
}

func incidentDocFromPayload(payload interface{}) (*incidents.IncidentDoc, bool) {
	switch doc := payload.(type) {
	case incidents.IncidentDoc:
		return &doc, true
	case *incidents.IncidentDoc:
		return doc, doc != nil
	default:
		return nil, false
	}
}
