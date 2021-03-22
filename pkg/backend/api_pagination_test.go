package backend

import (
	"testing"

	"github.com/alice-lg/alice-lg/pkg/api"
)

func TestApiRoutesPagination(t *testing.T) {
	routes := api.Routes{
		&api.Route{Id: "r01"},
		&api.Route{Id: "r02"},
		&api.Route{Id: "r03"},
		&api.Route{Id: "r04"},
		&api.Route{Id: "r05"},
		&api.Route{Id: "r06"},
		&api.Route{Id: "r07"},
		&api.Route{Id: "r08"},
		&api.Route{Id: "r09"},
		&api.Route{Id: "r10"},
	}

	paginated, pagination := apiPaginateRoutes(routes, 0, 8)

	if pagination.TotalPages != 2 {
		t.Error("Expected total pages to be 2, got:", pagination.TotalPages)
	}

	if pagination.TotalResults != 10 {
		t.Error("Expected total results to be 10, got:", pagination.TotalResults)
	}

	if pagination.Page != 0 {
		t.Error("Exptected current page to be 0, got:", pagination.Page)
	}

	// Check paginated slicing
	r := paginated[0]
	if r.Id != "r01" {
		t.Error("First route on page 0 should be r01, got:", r.Id)
	}

	r = paginated[len(paginated)-1]
	if r.Id != "r08" {
		t.Error("Last route should be r08, but got:", r.Id)
	}

	// Second page
	paginated, _ = apiPaginateRoutes(routes, 1, 8)
	if len(paginated) != 2 {
		t.Error("There should be 2 routes left on page 1, got:", len(paginated))
	}

	r = paginated[0]
	if r.Id != "r09" {
		t.Error("First route on page 1 should be r09, got:", r.Id)
	}

	r = paginated[len(paginated)-1]
	if r.Id != "r10" {
		t.Error("Last route should be r10, but got:", r.Id)
	}

	// Access out of bound page
	paginated, _ = apiPaginateRoutes(routes, 1000, 8)
	if len(paginated) > 0 {
		t.Error("There should be nothing on this page")
	}
}
