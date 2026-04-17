package tui

import (
	"context"

	"github.com/smasonuk/falken-core/pkg/falken"
)

type interactionPermissionRequest struct {
	Request  falken.PermissionRequest
	Response chan falken.PermissionResponse
}

type interactionPlanRequest struct {
	Request  falken.PlanApprovalRequest
	Response chan falken.PlanApprovalResponse
}

type SessionBridge struct {
	events             chan falken.Event
	permissionRequests chan interactionPermissionRequest
	planRequests       chan interactionPlanRequest
}

func NewSessionBridge() *SessionBridge {
	return &SessionBridge{
		events:             make(chan falken.Event, 256),
		permissionRequests: make(chan interactionPermissionRequest),
		planRequests:       make(chan interactionPlanRequest),
	}
}

func (b *SessionBridge) EventChannel() chan falken.Event {
	return b.events
}

func (b *SessionBridge) NextPermissionRequest() <-chan interactionPermissionRequest {
	return b.permissionRequests
}

func (b *SessionBridge) NextPlanRequest() <-chan interactionPlanRequest {
	return b.planRequests
}

func (b *SessionBridge) OnEvent(event falken.Event) {
	b.events <- event
}

func (b *SessionBridge) RequestPermission(ctx context.Context, req falken.PermissionRequest) (falken.PermissionResponse, error) {
	response := make(chan falken.PermissionResponse, 1)
	request := interactionPermissionRequest{Request: req, Response: response}

	select {
	case b.permissionRequests <- request:
	case <-ctx.Done():
		return falken.PermissionResponse{}, ctx.Err()
	}

	select {
	case result := <-response:
		return result, nil
	case <-ctx.Done():
		return falken.PermissionResponse{}, ctx.Err()
	}
}

func (b *SessionBridge) RequestPlanApproval(ctx context.Context, req falken.PlanApprovalRequest) (falken.PlanApprovalResponse, error) {
	response := make(chan falken.PlanApprovalResponse, 1)
	request := interactionPlanRequest{Request: req, Response: response}

	select {
	case b.planRequests <- request:
	case <-ctx.Done():
		return falken.PlanApprovalResponse{}, ctx.Err()
	}

	select {
	case result := <-response:
		return result, nil
	case <-ctx.Done():
		return falken.PlanApprovalResponse{}, ctx.Err()
	}
}

func (b *SessionBridge) OnSubmit(ctx context.Context, req falken.SubmitRequest) error {
	b.events <- falken.Event{
		Type:          falken.EventTypeWorkSubmitted,
		WorkSubmitted: &falken.WorkSubmittedEvent{Summary: req.Summary},
	}
	return nil
}
