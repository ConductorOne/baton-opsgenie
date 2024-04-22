package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	ogClient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/og"
	ogSchedule "github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
)

const (
	scheduleMember = "member"
	scheduleOnCall = "on-call"

	userParticipantType       = "user"
	teamParticipantType       = "team"
	escalationParticipantType = "escalation"
)

type scheduleResourceType struct {
	resourceType *v2.ResourceType
	config       *ogClient.Config
}

func (s *scheduleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return s.resourceType
}

// parses array of rotations and returns all teams and users that participate in rotations.
func parseRotations(rotation []og.Rotation) ([]string, []string) {
	var teams, users []string

	for _, r := range rotation {
		for _, p := range r.Participants {
			if p.Type == teamParticipantType {
				teams = append(teams, p.Id)
			} else if p.Type == userParticipantType {
				users = append(users, p.Id)
			}
		}
	}

	return teams, users
}

// scheduleResource creates a new connector resource for a OpsGenie Schedule.
func scheduleResource(schedule *ogSchedule.Schedule) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"schedule_id":   schedule.Id,
		"schedule_name": schedule.Name,
	}

	teams, users := parseRotations(schedule.Rotations)

	if len(teams) > 0 {
		profile["schedule_teams"] = scheduleParticipantsToInterfaceSlice(teams)
	}

	if len(users) > 0 {
		profile["schedule_users"] = scheduleParticipantsToInterfaceSlice(users)
	}

	resource, err := rs.NewGroupResource(
		schedule.Name,
		resourceTypeSchedule,
		schedule.Id,
		[]rs.GroupTraitOption{rs.WithGroupProfile(profile)},
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (s *scheduleResourceType) List(ctx context.Context, parentID *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	client, err := ogSchedule.NewClient(s.config)
	if err != nil {
		return nil, "", nil, err
	}

	expand := true
	req := &ogSchedule.ListRequest{
		BaseRequest: ogClient.BaseRequest{},
		Expand:      &expand,
	}
	schedules, err := client.List(ctx, req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("opsgenie-connector: failed to list schedules: %w", err)
	}

	var rv []*v2.Resource
	for _, schedule := range schedules.Schedule {
		scheduleCopy := schedule

		sr, err := scheduleResource(&scheduleCopy)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, sr)
	}

	return rv, "", nil, nil
}

func (s *scheduleResourceType) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	memberEntitlementOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser, resourceTypeTeam),
		ent.WithDisplayName(fmt.Sprintf("%s schedule %s", resource.DisplayName, scheduleMember)),
		ent.WithDescription(fmt.Sprintf("%s OpsGenie schedule %s", resource.DisplayName, scheduleMember)),
	}

	oncallEntitlementOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s schedule %s", resource.DisplayName, scheduleOnCall)),
		ent.WithDescription(fmt.Sprintf("%s OpsGenie schedule %s", resource.DisplayName, scheduleOnCall)),
	}

	rv = append(
		rv,
		ent.NewAssignmentEntitlement(resource, scheduleMember, memberEntitlementOptions...),
		ent.NewAssignmentEntitlement(resource, scheduleOnCall, oncallEntitlementOptions...),
	)

	return rv, "", nil, nil
}

func (s *scheduleResourceType) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	// parse resource profile to get schedule members (users or teams)
	groupTrait, err := rs.GetGroupTrait(resource)
	if err != nil {
		return nil, "", nil, err
	}

	users, ok := getProfileStringArray(groupTrait.Profile, "schedule_users")
	if !ok {
		l.Info("opsgenie-connector: no users found for schedule resource")
	}

	teams, ok := getProfileStringArray(groupTrait.Profile, "schedule_teams")
	if !ok {
		l.Info("opsgenie-connector: no teams found for schedule resource")
	}

	var rv []*v2.Grant

	// grant users and teams under schedule the member entitlement
	for _, u := range users {
		rv = append(rv, grant.NewGrant(
			resource,
			scheduleMember,
			&v2.ResourceId{
				ResourceType: resourceTypeUser.Id,
				Resource:     u,
			},
		))
	}

	for _, t := range teams {
		rv = append(rv, grant.NewGrant(
			resource,
			scheduleMember,
			&v2.ResourceId{
				ResourceType: resourceTypeTeam.Id,
				Resource:     t,
			},
			grant.WithAnnotation(
				&v2.GrantExpandable{
					EntitlementIds: []string{fmt.Sprintf("team:%s:%s", t, scheduleMember)},
				},
			),
		))
	}

	// grant users and teams under schedule the on-call entitlement
	client, err := ogSchedule.NewClient(s.config)
	if err != nil {
		return nil, "", nil, err
	}

	flat := false
	req := &ogSchedule.GetOnCallsRequest{
		BaseRequest:        ogClient.BaseRequest{},
		Flat:               &flat,
		ScheduleIdentifier: resource.DisplayName,
	}

	oncalls, err := client.GetOnCalls(ctx, req)
	if err != nil {
		return nil, "", nil, fmt.Errorf("opsgenie-connector: failed to list on-calls: %w", err)
	}

	for _, p := range oncalls.OnCallParticipants {
		var resourceType string
		var grantOptions []grant.GrantOption

		switch p.Type {
		case userParticipantType:
			resourceType = resourceTypeUser.Id
		case teamParticipantType:
			resourceType = resourceTypeTeam.Id
			grantOptions = append(
				grantOptions,
				grant.WithAnnotation(
					&v2.GrantExpandable{
						EntitlementIds: []string{fmt.Sprintf("team:%s:%s", p.Id, teamMemberEntitlement)},
					},
				),
			)
		case escalationParticipantType:
			resourceType = resourceTypeEscalation.Id
		default:
			return nil, "", nil, fmt.Errorf("opsgenie-connector: unknown participant type: %s", p.Type)
		}

		rv = append(rv, grant.NewGrant(
			resource,
			scheduleOnCall,
			&v2.ResourceId{
				ResourceType: resourceType,
				Resource:     p.Id,
			},
			grantOptions...,
		))
	}

	return rv, "", nil, nil
}

func scheduleBuilder(config *ogClient.Config) *scheduleResourceType {
	return &scheduleResourceType{
		resourceType: resourceTypeSchedule,
		config:       config,
	}
}
