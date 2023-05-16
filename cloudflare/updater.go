package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

type Handler struct {
	*cloudflare.API
}

func (h *Handler) Update(domain string, address string) error {
	zoneName := domain
	if strings.Count(zoneName, ".") > 1 {
		ps := strings.Split(domain, ".")
		zoneName = fmt.Sprintf("%s.%s", ps[len(ps)-2], ps[len(ps)-1])
	}

	zoneID, err := h.API.ZoneIDByName(zoneName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	rc := cloudflare.ZoneIdentifier(zoneID)
	params := cloudflare.ListDNSRecordsParams{Name: domain}
	records, _, err := h.API.ListDNSRecords(ctx, rc, params)
	if err != nil {
		return err
	}

	if address == "" {
		if len(records) > 0 {
			return h.API.DeleteDNSRecord(ctx, rc, records[0].ID)
		}
		return nil
	}

	if len(records) == 0 {
		params := cloudflare.CreateDNSRecordParams{
			Type:    "A",
			Name:    domain,
			Content: address,
			Proxied: cloudflare.BoolPtr(false),
			TTL:     60,
		}
		_, err := h.API.CreateDNSRecord(ctx, rc, params)
		return err
	} else {
		record := records[0]
		if record.Content == address {
			return nil
		}
		params := cloudflare.UpdateDNSRecordParams{
			ID:      record.ID,
			Content: address,
		}
		_, err = h.API.UpdateDNSRecord(ctx, rc, params)
		return err
	}
}
