package connector

import (
	"context"
	"io"
	"math"
	"net/http"
	"time"

	zaphook "github.com/Sytten/logrus-zap-hook"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	ogclient "github.com/opsgenie/opsgenie-go-sdk-v2/client"
	user "github.com/opsgenie/opsgenie-go-sdk-v2/user"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

var (
	resourceTypeRole = &v2.ResourceType{
		Id:          "role",
		DisplayName: "Role",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_ROLE},
	}
	resourceTypeTeam = &v2.ResourceType{
		Id:          "team",
		DisplayName: "Team",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}
	resourceTypeUser = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_USER,
		},
		Annotations: annotationsForUserResourceType(),
	}
	resourceTypeSchedule = &v2.ResourceType{
		Id:          "schedule",
		DisplayName: "Schedule",
		Traits: []v2.ResourceType_Trait{
			v2.ResourceType_TRAIT_GROUP,
		},
	}
)

type Opsgenie struct {
	config *ogclient.Config
	apiKey string
}

func New(ctx context.Context, apiKey string) (*Opsgenie, error) {
	l := ctxzap.Extract(ctx)
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, l))
	if err != nil {
		return nil, err
	}

	// OpsGenie client takes a logrus logger, but we use zap.
	logger := logrus.New()
	logger.ReportCaller = true   // So Zap reports the right caller
	logger.SetOutput(io.Discard) // Prevent logrus from writing its logs

	hook, err := zaphook.NewZapHook(l)
	if err != nil {
		return nil, err
	}

	logger.Hooks.Add(hook)

	clientConfig := &ogclient.Config{
		ApiKey:     apiKey,
		HttpClient: httpClient,
		Logger:     logger,
		RetryCount: 20,
		Backoff: func(_, _ time.Duration, attemptNum int, resp *http.Response) time.Duration {
			// exponential backoff - more information about rate limits in OpsGenie here: https://docs.opsgenie.com/docs/api-rate-limiting
			exp := math.Pow(2, float64(attemptNum))
			t := time.Duration(200) * time.Millisecond

			l.Debug("retrying in ", zap.Duration("retry_in", t*time.Duration(exp)))
			return t * time.Duration(exp)
		},
	}

	rv := &Opsgenie{
		apiKey: apiKey,
		config: clientConfig,
	}

	return rv, nil
}

func (c *Opsgenie) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Opsgenie",
		Description: "Connector syncing Opsgenie users, teams and roles to Baton",
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
		scheduleBuilder(c.config),
	}
}
