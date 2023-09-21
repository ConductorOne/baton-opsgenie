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
		annos := &v2.V1Identifier{
			Id: t.Id,
		}
		profile := teamProfile(ctx, t)
		groupTrait := []res.GroupTraitOption{res.WithGroupProfile(profile)}
		groupResource, err := res.NewGroupResource(t.Name, resourceTypeTeam, t.Id, groupTrait, res.WithAnnotation(annos))
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, groupResource)
	}
	return rv, "", nil, nil
}

func (o *teamResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var annos annotations.Annotations
	annos.Update(&v2.V1Identifier{
		Id: V1MembershipEntitlementID(resource.Id.Resource),
	})
	member := ent.NewAssignmentEntitlement(resource, teamMemberEntitlement, ent.WithGrantableTo(resourceTypeUser))
	member.Description = fmt.Sprintf("Is member of the %s team in Opsgenie", resource.DisplayName)
	member.Annotations = annos
	member.DisplayName = fmt.Sprintf("%s Team Member", resource.DisplayName)
	return []*v2.Entitlement{member}, "", nil, nil
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
			v1Identifier := &v2.V1Identifier{
				Id: V1GrantID(V1MembershipEntitlementID(resource.Id.Resource), member.User.ID),
			}
			gmID, err := res.NewResourceID(resourceTypeUser, member.User.ID)
			if err != nil {
				return nil, "", nil, err
			}
			grant := grant.NewGrant(resource, teamMemberEntitlement, gmID, grant.WithAnnotation(v1Identifier))
			rv = append(rv, grant)
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

func teamProfile(ctx context.Context, team oteam.ListedTeams) map[string]interface{} {
	profile := make(map[string]interface{})
	profile["team_id"] = team.Id
	profile["team_name"] = team.Name
	profile["team_description"] = team.Description
	return profile
}
