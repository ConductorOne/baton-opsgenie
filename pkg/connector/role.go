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
		annos := &v2.V1Identifier{
			Id: role.Id,
		}
		profile := roleProfile(ctx, role.Name, role.Id)
		roleTrait := []res.RoleTraitOption{res.WithRoleProfile(profile)}
		roleResource, err := res.NewRoleResource(role.Name, resourceTypeRole, role.Id, roleTrait, res.WithAnnotation(annos))
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, roleResource)
	}

	// Adds the default roles not returned by the custom roles endpoint
	for roleName, id := range defaultRoles {
		annos := &v2.V1Identifier{
			Id: id,
		}
		profile := roleProfile(ctx, roleName, id)
		roleTrait := []res.RoleTraitOption{res.WithRoleProfile(profile)}
		roleResource, err := res.NewRoleResource(roleName, resourceTypeRole, id, roleTrait, res.WithAnnotation(annos))
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, roleResource)
	}
	return rv, "", nil, nil
}

func (o *roleResourceType) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var annos annotations.Annotations
	annos.Update(&v2.V1Identifier{
		Id: V1MembershipEntitlementID(resource.Id.Resource),
	})
	member := ent.NewAssignmentEntitlement(resource, roleMemberEntitlement, ent.WithGrantableTo(resourceTypeUser))
	member.Description = fmt.Sprintf("Has the %s role in Opsgenie", resource.DisplayName)
	member.Annotations = annos
	member.DisplayName = fmt.Sprintf("%s Role Member", resource.DisplayName)
	return []*v2.Entitlement{member}, "", nil, nil
}

func (o *roleResourceType) Grants(ctx context.Context, resource *v2.Resource, pt *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pt.Token)
	if err != nil {
		return nil, "", nil, err
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: resource.Id.ResourceType,
			ResourceID:     resource.Id.Resource,
		})
	}

	userClient, err := user.NewClient(o.config)
	if err != nil {
		return nil, "", nil, err
	}

	users, err := userClient.List(ctx, &user.ListRequest{Limit: int(100), Offset: strToInt(bag.PageToken())})
	if err != nil {
		return nil, "", nil, err
	}
	var rv []*v2.Grant
	for _, user := range users.Users {
		if user.Role.RoleName == resource.DisplayName {
			v1Identifier := &v2.V1Identifier{
				Id: V1GrantID(V1MembershipEntitlementID(resource.Id.Resource), user.Id),
			}
			rID, err := res.NewResourceID(resourceTypeUser, user.Id)
			if err != nil {
				return nil, "", nil, err
			}
			grant := grant.NewGrant(resource, roleMemberEntitlement, rID, grant.WithAnnotation(v1Identifier))
			rv = append(rv, grant)
		}
	}

	nextPage, err := bag.NextToken(users.Paging.Next)
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

func roleProfile(ctx context.Context, roleName string, roleId string) map[string]interface{} {
	profile := make(map[string]interface{})
	profile["role_id"] = roleId
	profile["role_name"] = roleName
	return profile
}
