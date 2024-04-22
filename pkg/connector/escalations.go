package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	res "github.com/conductorone/baton-sdk/pkg/types/resource"
	ogclient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	ogEscalation "github.com/opsgenie/opsgenie-go-sdk-v2/escalation"
)

const (
	escalationMemberEntitlement = "member"
)

type escalationResourceType struct {
	resourceType *v2.ResourceType
	config       *ogclient.Config
}

func (o *escalationResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func escalationResource(ctx context.Context, escalation ogEscalation.Escalation) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"escalation_id":          escalation.Id,
		"escalation_name":        escalation.Name,
		"escalation_description": escalation.Description,
	}

	groupTraitOptions := []res.GroupTraitOption{
		res.WithGroupProfile(profile),
	}

	resource, err := res.NewGroupResource(
		escalation.Name,
		resourceTypeEscalation,
		escalation.Id,
		groupTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *escalationResourceType) List(ctx context.Context, resourceId *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	escalationClient, err := ogEscalation.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	escalations, err := escalationClient.List(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, 0)
	for _, t := range escalations.Escalations {
		tr, err := escalationResource(ctx, t)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, tr)
	}

	return rv, "", nil, nil
}

func (o *escalationResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeEscalation),
		ent.WithDisplayName(fmt.Sprintf("%s Escalation Member", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("Is member of the %s escalation in Opsgenie", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		escalationMemberEntitlement,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (o *escalationResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func escalationBuilder(config *ogclient.Config) *escalationResourceType {
	return &escalationResourceType{
		resourceType: resourceTypeEscalation,
		config:       config,
	}
}
