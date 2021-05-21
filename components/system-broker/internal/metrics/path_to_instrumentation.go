package metrics

import "net/http"

type Path struct {
	PathTemplate string
	HTTPMethods  string
}

var PathToInstrumentationMapping map[Path]func(c *Collector, handler http.Handler) http.HandlerFunc

func MapPathToInstrumentation(rootAPI string) {
	PathToInstrumentationMapping = map[Path]func(c *Collector, handler http.Handler) http.HandlerFunc{
		Path{
			PathTemplate: rootAPI + "/v2/catalog",
			HTTPMethods:  "GET",
		}: func(c *Collector, handler http.Handler) http.HandlerFunc {
			return c.CatalogHandlerWithInstrumentation(handler)
		},
		Path{
			PathTemplate: rootAPI + "/v2/service_instances/{instance_id}",
			HTTPMethods:  "PUT",
		}: func(c *Collector, handler http.Handler) http.HandlerFunc {
			return c.ProvisionHandlerWithInstrumentation(handler)
		},
		Path{
			PathTemplate: rootAPI + "/v2/service_instances/{instance_id}",
			HTTPMethods:  "DELETE",
		}: func(c *Collector, handler http.Handler) http.HandlerFunc {
			return c.DeprovosionHandlerWithInstrumentation(handler)
		},
		Path{
			PathTemplate: rootAPI + "/v2/service_instances/{instance_id}/service_bindings/{binding_id}",
			HTTPMethods:  "PUT",
		}: func(c *Collector, handler http.Handler) http.HandlerFunc {
			return c.BindHandlerWithInstrumentation(handler)
		},
		Path{
			PathTemplate: rootAPI + "/v2/service_instances/{instance_id}/service_bindings/{binding_id}",
			HTTPMethods:  "DELETE",
		}: func(c *Collector, handler http.Handler) http.HandlerFunc {
			return c.UnbindHandlerWithInstrumentation(handler)
		},
	}
}
