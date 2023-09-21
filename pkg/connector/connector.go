package connector

import (
	"context"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	ogclient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	user "github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

var (
	resourceTypeRole = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
		Annotations: v1AnnotationsForResourceType("role"),
	}
	resourceTypeTeam = &v2.ResourceType{
		Id:          "team",
		DisplayName: "Team",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
		Annotations: v1AnnotationsForResourceType("team"),
	}
	resourceTypeUser = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
		Annotations: v1AnnotationsForResourceType("user"),
	}
)

type Config struct {
	ApiKey string
}
type Opsgenie struct {
	config *ogclient.Config
	apiKey string
}

func New(ctx context.Context, config Config) (*Opsgenie, error) {
	l := ctxzap.Extract(ctx)
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, l))
	if err != nil {
		return nil, err
	}

	clientConfig := &ogclient.Config{
		ApiKey:     config.ApiKey,
		HttpClient: httpClient,
	}

	rv := &Opsgenie{
		apiKey: config.ApiKey,
		config: clientConfig,
	}
	return rv, nil
}

func (c *Opsgenie) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	_, err := c.Validate(ctx)
	if err != nil {
		return nil, err
	}

	return &v2.ConnectorMetadata{
		DisplayName: "Opsgenie",
	}, nil
}

func (c *Opsgenie) Validate(ctx context.Context) (annotations.Annotations, error) {
	userClient, err := user.NewClient(c.config)
	if err != nil {
		return nil, err
	}

	_, err = userClient.List(ctx, &user.ListRequest{Limit: 1, Offset: 0})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Opsgenie) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

func (c *Opsgenie) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		teamBuilder(c.config),
		roleBuilder(c.config),
		userBuilder(c.config),
	}
}
