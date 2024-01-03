package connector

import (
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"google.golang.org/protobuf/types/known/structpb"
)

const ResourcesPageSize = 100

func strToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func annotationsForUserResourceType() annotations.Annotations {
	annos := annotations.Annotations{}
	annos.Update(&v2.SkipEntitlementsAndGrants{})
	return annos
}

func getProfileStringArray(profile *structpb.Struct, k string) ([]string, bool) {
	var values []string
	if profile == nil {
		return nil, false
	}

	v, ok := profile.Fields[k]
	if !ok {
		return nil, false
	}

	s, ok := v.Kind.(*structpb.Value_ListValue)
	if !ok {
		return nil, false
	}

	for _, v := range s.ListValue.Values {
		if strVal := v.GetStringValue(); strVal != "" {
			values = append(values, strVal)
		}
	}

	return values, true
}

func scheduleParticipantsToInterfaceSlice(p []string) []interface{} {
	var i []interface{}
	for _, v := range p {
		i = append(i, v)
	}
	return i
}
