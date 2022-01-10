package store

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"encoding/json"
	"io/ioutil"

	"github.com/alice-lg/alice-lg/pkg/api"
	"github.com/alice-lg/alice-lg/pkg/config"
	"github.com/alice-lg/alice-lg/pkg/store/backends/memory"
)

func loadTestRoutesResponse() *api.RoutesResponse {
	file, err := os.Open("testdata/routes_response.json")
	if err != nil {
		log.Panic("could not load test data:", err)
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Panic("could not read test data:", err)
	}
	response := &api.RoutesResponse{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Panic("could not unmarshal response test data:", err)
	}
	return response
}

func importRoutes(
	s *RoutesStore,
	src *config.SourceConfig,
	res *api.RoutesResponse,
) error {
	ctx := context.Background()

	// Prepare imported routes for lookup
	imported := s.routesToLookupRoutes(ctx, "imported", src, res.Imported)
	filtered := s.routesToLookupRoutes(ctx, "filtered", src, res.Filtered)
	lookupRoutes := append(imported, filtered...)

	if err := s.backend.SetRoutes(ctx, src.ID, lookupRoutes); err != nil {
		return err
	}

	return s.sources.RefreshSuccess(src.ID)
}

//
// Route Store Tests
//
func makeTestRoutesStore() *RoutesStore {
	neighborsStore := makeTestNeighborsStore()
	be := memory.NewRoutesBackend()

	cfg := &config.Config{
		Server: config.ServerConfig{
			RoutesStoreRefreshInterval: 1,
		},
		Sources: []*config.SourceConfig{
			{
				ID:   "rs1",
				Name: "rs1",
			},
			{
				ID:   "rs2",
				Name: "rs2",
			},
		},
	}
	rs1 := loadTestRoutesResponse()
	s := NewRoutesStore(neighborsStore, cfg, be)
	if err := importRoutes(s, cfg.Sources[0], rs1); err != nil {
		log.Panic(err)
	}
	return s
}

// Check for presence of network in result set
func testCheckPrefixesPresence(prefixes, resultset []string, t *testing.T) {
	// Check prefixes
	presence := map[string]bool{}
	for _, prefix := range prefixes {
		presence[prefix] = false
	}

	for _, prefix := range resultset {
		// Check if prefixes are all accounted for
		for net := range presence {
			if prefix == net {
				presence[net] = true
			}
		}
	}

	for net, present := range presence {
		if present == false {
			t.Error(net, "not found in result set")
		}
	}
}

func TestRoutesStoreStats(t *testing.T) {

	store := makeTestRoutesStore()
	stats := store.Stats()

	// Check total routes
	// There should be 8 imported, and 1 filtered route
	if stats.TotalRoutes.Imported != 8 {
		t.Error(
			"expected 8 imported routes, got:",
			stats.TotalRoutes.Imported,
		)
	}

	if stats.TotalRoutes.Filtered != 1 {
		t.Error(
			"expected 1 filtered route, got:",
			stats.TotalRoutes.Filtered,
		)
	}
}

func TestLookupPrefix(t *testing.T) {
	store := makeTestRoutesStore()
	query := "193.200."

	results, err := store.LookupPrefix(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) == 0 {
		t.Error("Expected lookup results. None present.")
		return
	}

	// Check results
	for _, prefix := range results {
		if strings.HasPrefix(prefix.Network, query) == false {
			t.Error(
				"All network addresses should start with the",
				"queried prefix",
			)
		}
	}
}

func TestLookupPrefixForNeighbors(t *testing.T) {
	// Construct a neighbors lookup result
	neighbors := api.NeighborsLookupResults{
		"rs1": api.Neighbors{
			&api.Neighbor{
				ID: "ID163_AS31078",
			},
		},
	}

	store := makeTestRoutesStore()

	// Query
	results, err := store.LookupPrefixForNeighbors(context.Background(), neighbors)
	if err != nil {
		t.Fatal(err)
	}

	// We should have retrived 8 prefixes,
	if len(results) != 8 {
		t.Error("Expected result lenght: 8, got:", len(results))
	}

	presence := []string{
		"193.200.230.0/24", "193.34.24.0/22", "31.220.136.0/21",
	}

	resultset := []string{}
	for _, prefix := range results {
		resultset = append(resultset, prefix.Network)
	}

	testCheckPrefixesPresence(presence, resultset, t)
}