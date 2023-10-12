package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	res "github.com/conductorone/baton-sdk/pkg/types/resource"
	ogclient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	oteam "github.com/opsgenie/opsgenie-go-sdk-v2/team"
)

const (
	teamMemberEntitlement = "member"
	idIdentifierType      = 1
)

type teamResourceType struct {
	resourceType *v2.ResourceType
	config       *ogclient.Config
}

func (o *teamResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func teamResource(ctx context.Context, team oteam.ListedTeams) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"team_id":          team.Id,
		"team_name":        team.Name,
		"team_description": team.Description,
	}

	groupTraitOptions := []res.GroupTraitOption{
		res.WithGroupProfile(profile),
	}

	resource, err := res.NewGroupResource(
		team.Name,
		resourceTypeTeam,
		team.Id,
		groupTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *teamResourceType) List(ctx context.Context, resourceId *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	teamClient, err := oteam.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	teams, err := teamClient.List(ctx, &oteam.ListTeamRequest{BaseRequest: ogclient.BaseRequest{}})
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, 0)
	for _, t := range teams.Teams {
		tr, err := teamResource(ctx, t)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, tr)
	}

	return rv, "", nil, nil
}

func (o *teamResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s Team Member", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("Is member of the %s team in Opsgenie", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		teamMemberEntitlement,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (o *teamResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var rv []*v2.Grant
	teamClient, err := oteam.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	teams, err := teamClient.List(ctx, &oteam.ListTeamRequest{BaseRequest: ogclient.BaseRequest{}})
	if err != nil {
		return nil, "", nil, err
	}

	for _, team := range teams.Teams {
		teamWithMembers, err := teamClient.Get(ctx, &oteam.GetTeamRequest{BaseRequest: ogclient.BaseRequest{}, IdentifierValue: team.Id, IdentifierType: oteam.Identifier(idIdentifierType)})
		if err != nil {
			return nil, "", nil, err
		}

		for _, member := range teamWithMembers.Members {
			rv = append(rv, grant.NewGrant(
				resource,
				teamMemberEntitlement,
				&v2.ResourceId{
					ResourceType: resourceTypeUser.Id,
					Resource:     member.User.ID,
				},
			))
		}
	}

	return rv, "", nil, nil
}

func teamBuilder(config *ogclient.Config) *teamResourceType {
	return &teamResourceType{
		resourceType: resourceTypeTeam,
		config:       config,
	}
}
