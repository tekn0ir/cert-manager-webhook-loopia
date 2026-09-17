package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	loopia "github.com/jonlil/loopia-go"
)

// TestLoopiaAPIAuthenticationSignature performs a live check against the Loopia
// XMLRPC API using the same client code path as loopiaDNSProviderSolver.Present
// and loopiaDNSProviderSolver.CleanUp.
//
// It verifies that the authentication arguments sent by the Loopia client are
// accepted by the API and that the wildcard flow of the solver works, i.e. two
// TXT records presented in the same '_acme-challenge' sub domain are both
// created, resolvable by id and removable again.
//
// It is skipped unless real Loopia API credentials are supplied:
//
//	LOOPIA_USERNAME=... LOOPIA_PASSWORD=... LOOPIA_TEST_ZONE=example.com \
//	  go test -run TestLoopiaAPIAuthenticationSignature -v .
func TestLoopiaAPIAuthenticationSignature(t *testing.T) {
	username := os.Getenv("LOOPIA_USERNAME")
	password := os.Getenv("LOOPIA_PASSWORD")
	zone := os.Getenv("LOOPIA_TEST_ZONE")
	if username == "" || password == "" || zone == "" {
		t.Skip("LOOPIA_USERNAME, LOOPIA_PASSWORD and LOOPIA_TEST_ZONE must be set to run the live Loopia API test")
	}

	client, err := loopia.New(username, password)
	if err != nil {
		t.Fatalf("could not initialize Loopia client: %v", err)
	}

	// Any call requires Loopia to accept the authentication arguments the client sends.
	if _, err := client.GetZoneRecords(zone, "@"); err != nil {
		t.Fatalf("GetZoneRecords failed: %v", err)
	}

	subdomain := fmt.Sprintf("_acme-challenge-cert-%d", time.Now().Unix())
	records := []*loopia.Record{
		{TTL: LoopiaMinTtl, Type: "TXT", Value: "cert-manager-dns01-domain", Priority: 0},
		{TTL: LoopiaMinTtl, Type: "TXT", Value: "cert-manager-dns01-wildcard", Priority: 0},
	}

	for i, record := range records {
		if err := client.AddZoneRecord(zone, subdomain, record); err != nil {
			t.Fatalf("unable to create txt-record %d: %v", i, err)
		}
		if record.ID == 0 {
			t.Fatalf("txt-record %d was not created, no record_id was returned", i)
		}
	}

	zoneRecords, err := client.GetZoneRecords(zone, subdomain)
	if err != nil {
		t.Fatalf("unable to get zone records: %v", err)
	}
	for _, record := range records {
		found := false
		for _, zoneRecord := range zoneRecords {
			if zoneRecord.ID == record.ID && zoneRecord.Value == record.Value {
				found = true
			}
		}
		if !found {
			t.Errorf("txt-record with id %d and value %q was not returned by getZoneRecords", record.ID, record.Value)
		}
	}

	for _, record := range records {
		if _, err := client.RemoveZoneRecord(zone, subdomain, record.ID); err != nil {
			t.Errorf("unable to delete TXT record: %v", err)
		}
	}
	if _, err := client.RemoveSubDomain(zone, subdomain); err != nil {
		t.Errorf("unable to remove subdomain: %v", err)
	}
}
