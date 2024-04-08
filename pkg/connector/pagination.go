package connector

import (
	"net/url"
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func parsePageToken(i string, resourceID *v2.ResourceId) (*pagination.Bag, int, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(i)
	if err != nil {
		return nil, 0, err
	}

	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceTypeID: resourceID.ResourceType,
			ResourceID:     resourceID.Resource,
		})
	}

	page, err := convertPageToken(b.PageToken())
	if err != nil {
		return nil, 0, err
	}

	return b, page, nil
}

func convertPageToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}

	page, err := strconv.ParseInt(token, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(page), nil
}

func handleNextPage(bag *pagination.Bag, nextLink string) (string, error) {
	if nextLink == "" {
		return "", nil
	}

	nextUrl, err := url.Parse(nextLink)
	if err != nil {
		return "", err
	}

	offset := nextUrl.Query().Get("offset")
	if offset == "" {
		return "", nil
	}

	pageToken, err := bag.NextToken(offset)
	if err != nil {
		return "", err
	}

	return pageToken, nil
}
