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

func userResource(ctx context.Context, user user.User) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"full_name": user.FullName,
		"time_zone": user.TimeZone,
		"blocked":   user.Blocked,
		"verified":  user.Verified,
		"email":     user.Username,
	}

	userTraitOptions := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithEmail(user.Username, true),
		resource.WithStatus(v2.UserTrait_Status_STATUS_ENABLED),
	}

	resource, err := resource.NewUserResource(
		user.FullName,
		resourceTypeUser,
		user.Id,
		userTraitOptions,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *userResourceType) List(ctx context.Context, _ *v2.ResourceId, pt *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, offset, err := parsePageToken(pt.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
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

	rv := make([]*v2.Resource, 0)
	for _, user := range users.Users {
		userCopy := user

		ur, err := userResource(ctx, userCopy)
		if err != nil {
			return nil, "", nil, err
		}

		rv = append(rv, ur)
	}

	nextPage, err := handleNextPage(bag, users.Paging.Next)
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
