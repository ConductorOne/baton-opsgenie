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
	custom_role "github.com/opsgenie/opsgenie-go-sdk-v2/custom_user_role"
	user "github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

var defaultRoles = map[string]string{
	"Admin":       "2DonQtHeOOxwfe5muls9Cx18fzL",
	"User":        "2DonSZzqzbnAWBeWhKl6lNSTTzh",
	"Owner":       "2Dp2qL0NvHVku3cghi6rJSsvfXJ",
	"Stakeholder": "2Dp2txbavgGagm2sFtl7voULpf2",
}

const (
	roleMemberEntitlement = "member"
)

type roleResourceType struct {
	resourceType *v2.ResourceType
	config       *ogclient.Config
}

func (o *roleResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func roleResource(ctx context.Context, roleName, roleId string) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"role_id":   roleId,
		"role_name": roleName,
	}

	roleTraitOptions := []res.RoleTraitOption{
		res.WithRoleProfile(profile),
	}

	resource, err := res.NewRoleResource(
		roleName,
		resourceTypeRole,
		roleId,
		roleTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *roleResourceType) List(ctx context.Context, _ *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	crClient, err := custom_role.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	roles, err := crClient.List(ctx, &custom_role.ListRequest{BaseRequest: ogclient.BaseRequest{}})
	if err != nil {
		return nil, "", nil, err
	}

	rv := make([]*v2.Resource, 0)
	for _, role := range roles.CustomUserRoles {
		rr, err := roleResource(ctx, role.Name, role.Id)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	// Adds the default roles not returned by the custom roles endpoint
	for roleName, id := range defaultRoles {
		rr, err := roleResource(ctx, roleName, id)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, rr)
	}

	return rv, "", nil, nil
}

func (o *roleResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assignmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(resourceTypeUser),
		ent.WithDisplayName(fmt.Sprintf("%s Role Member", resource.DisplayName)),
		ent.WithDescription(fmt.Sprintf("Has the %s role in Opsgenie", resource.DisplayName)),
	}

	rv = append(rv, ent.NewAssignmentEntitlement(
		resource,
		roleMemberEntitlement,
		assignmentOptions...,
	))

	return rv, "", nil, nil
}

func (o *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pt.Token, resource.Id)
	if err != nil {
		return nil, "", nil, err
	}

	userClient, err := user.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	users, err := userClient.List(
		ctx,
		&user.ListRequest{
			Limit:  ResourcesPageSize,
			Offset: offset,
		},
	)
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	for _, user := range users.Users {
		if user.Role.RoleName == resource.DisplayName {
			rv = append(rv, grant.NewGrant(
				resource,
				roleMemberEntitlement,
				&v2.ResourceId{
					ResourceType: resourceTypeUser.Id,
					Resource:     user.Id,
				},
			))
		}
	}

	nextPage, err := handleNextPage(bag, users.Paging.Next)
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPage, nil, nil
}

func roleBuilder(config *ogclient.Config) *roleResourceType {
	return &roleResourceType{
		resourceType: resourceTypeRole,
		config:       config,
	}
}
