package main

import (
	"os"
	"strconv"
	"testing"
	"time"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	// Environment variable holding the name of the zone to test, ex: example.com however this needs to be a zone you have control over in Loopia.
	// This needs to be set before running the test.
	zone = os.Getenv("TEST_ZONE_NAME")

	// Environment variable for enabling the strict mode testing.
	strictmodeenv = os.Getenv("TEST_STRICT_MODE")

	// Environment variable for overriding how long to wait for a TXT-record to appear or disappear from the public DNS, ex: 20m.
	propagationlimitenv = os.Getenv("TEST_PROPAGATION_LIMIT")
)

// Loopia writes a record to the zone immediately, but its nameservers can take a long time to publish it.
// A TXT-record has been measured to become visible on ns1.loopia.se and ns2.loopia.se between 20 and 45 minutes
// after addZoneRecord returned OK, so the default of two minutes of the test fixture is far too short.
const defaultPropagationLimit = time.Minute * 60

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a sniplet of valid configuration that should be included on the ChallengeRequest passed as part of the test cases.
	// The test fixture also starts a complete Kubernetes control plane, which requires the envtest binaries (etcd, kube-apiserver and kubectl) to be available on the PATH or through KUBEBUILDER_ASSETS, there is a script supplied that downloads them in testdata/scripts.

	var strictmode, err = strconv.ParseBool(strictmodeenv)
	if err != nil {
		strictmode = false
	}

	propagationLimit := defaultPropagationLimit
	if propagationlimitenv != "" {
		propagationLimit, err = time.ParseDuration(propagationlimitenv)
		if err != nil {
			t.Fatalf("TEST_PROPAGATION_LIMIT is not a valid duration: %v", err)
		}
	}

	solver := &loopiaDNSProviderSolver{}
	fixture := acmetest.NewFixture(solver,
		acmetest.SetStrict(strictmode),
		acmetest.SetResolvedZone(zone),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetManifestPath("testdata/loopia"),
		acmetest.SetPollInterval(time.Second*60),
		acmetest.SetPropagationLimit(propagationLimit),
	)

	fixture.RunConformance(t)
}
