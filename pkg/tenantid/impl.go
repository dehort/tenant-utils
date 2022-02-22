package tenantid

import (
	"context"
	"fmt"
)

type translatorImpl struct {
	client      HttpRequestDoer
	serviceHost string
}

func (this *translatorImpl) EANToOrgID(ctx context.Context, ean string) (orgId string, err error) {
	return "", fmt.Errorf("Not implemented")
}

func (this *translatorImpl) OrgIDToEAN(ctx context.Context, orgId string) (ean *string, err error) {
	return nil, fmt.Errorf("Not implemented")
}
