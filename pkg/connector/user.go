package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	resource "github.com/conductorone/baton-sdk/pkg/types/resource"
	ogclient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	user "github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

type userResourceType struct {
	resourceType *v2.ResourceType
	config       *ogclient.Config
}

func (o *userResourceType) ResourceType(_ context.Context) *v2.ResourceType {
	return o.resourceType
}

func (o *userResourceType) List(ctx context.Context, _ *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag := &pagination.Bag{}
	err := bag.Unmarshal(pt.Token)
	if err != nil {
		return nil, "", nil, err
	}

	if bag.Current() == nil {
		bag.Push(pagination.PageState{
			ResourceTypeID: resourceTypeUser.Id,
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

	rv := make([]*v2.Resource, 0)
	for _, user := range users.Users {
		annos := &v2.V1Identifier{
			Id: user.Id,
		}
		profile := userProfile(ctx, user)
		userTrait := []resource.UserTraitOption{resource.WithUserProfile(profile), resource.WithEmail(user.Username, true), resource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED)}
		userResource, err := resource.NewUserResource(user.FullName, resourceTypeUser, user.Id, userTrait, resource.WithAnnotation(annos))
		if err != nil {
			return nil, "", nil, err
		}
		rv = append(rv, userResource)
	}

	nextPage, err := bag.NextToken(users.Paging.Next)
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPage, nil, nil
}

func (o *userResourceType) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *userResourceType) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func userBuilder(config *ogclient.Config) *userResourceType {
	return &userResourceType{
		resourceType: resourceTypeUser,
		config:       config,
	}
}

func userProfile(ctx context.Context, user user.User) map[string]interface{} {
	profile := make(map[string]interface{})
	profile["full_name"] = user.FullName
	profile["time_zone"] = user.TimeZone
	profile["blocked"] = user.Blocked
	profile["verified"] = user.Verified
	profile["email"] = user.Username
	return profile
}
