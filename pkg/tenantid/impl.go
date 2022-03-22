package tenantid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type batchTranslatorImpl struct {
	client      HttpRequestDoer
	serviceHost string
}

func (this *batchTranslatorImpl) batchRequest(ctx context.Context, operation string, path string, params []string) (map[string]string, error) {
	ctx = setOperation(ctx, operation)
	url := fmt.Sprintf("%s%s", this.serviceHost, path)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	r, err := this.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending HTTP request: %w", err)
	}

	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code %d", r.StatusCode)
	}

	var resp map[string]string

	err = json.NewDecoder(r.Body).Decode(&resp)
	if err != nil {
		return nil, fmt.Errorf("Error decoding HTTP response: %w", err)
	}

	return resp, nil
}

func (this *batchTranslatorImpl) EANsToOrgIDs(ctx context.Context, eans []string) ([]TranslationResult, error) {
	resp, err := this.batchRequest(ctx, "eans_to_org_ids", "/internal/orgIds", eans)
	if err != nil {
		return nil, err
	}

	results := make([]TranslationResult, len(eans))

	for i, ean := range eans {
		if value, ok := resp[ean]; ok {
			results[i] = newTranslationResult(value, stringRef(ean))
		} else {
			results[i] = TranslationResult{
				EAN: stringRef(ean),
				Err: &TenantNotFoundError{
					msg: fmt.Sprintf("Tenant not found. EAN: %s", ean),
				},
			}
		}
	}

	return results, nil
}

func (this *batchTranslatorImpl) OrgIDsToEANs(ctx context.Context, orgIDs []string) ([]TranslationResult, error) {
	resp, err := this.batchRequest(ctx, "org_ids_to_eans", "/internal/ebsNumbers", orgIDs)
	if err != nil {
		return nil, err
	}

	results := make([]TranslationResult, len(orgIDs))

	for i, orgID := range orgIDs {
		if value, ok := resp[orgID]; ok {
			results[i] = newTranslationResult(orgID, stringRef(value))
		} else {
			results[i] = newTranslationResult(orgID, nil)
		}
	}

	return results, nil
}

func newTranslationResult(orgID string, ean *string) TranslationResult {
	return TranslationResult{
		OrgID: orgID,
		EAN:   ean,
		Err:   nil,
	}
}
