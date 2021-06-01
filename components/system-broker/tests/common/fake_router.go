package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type GraphqlQueryKey struct {
	Type string
	Name string
}

type ConfigRequestBody struct {
	GraphqlQueryKey
	Response interface{}
}

type GqlFakeRouter struct {
	m              sync.RWMutex
	ResponseConfig map[GraphqlQueryKey][]interface{}
	Schema         *ast.Schema
}

func NewGqlFakeRouter(schemaName, path string) (*GqlFakeRouter, error) {
	schemaContent, err := getFileContent(path)
	if err != nil {
		return nil, err
	}

	schema, gqlErr := gqlparser.LoadSchema(&ast.Source{
		Name:  schemaName,
		Input: schemaContent,
	})
	if gqlErr != nil {
		return nil, errors.New(gqlErr.Error())
	}

	return &GqlFakeRouter{
		ResponseConfig: make(map[GraphqlQueryKey][]interface{}),
		Schema:         schema,
	}, nil
}

func (g *GqlFakeRouter) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/config", g.configHandler)
	mux.HandleFunc("/graphql", g.graphqlHandler)
	mux.HandleFunc("/verify", g.verifyHandler)
	mux.HandleFunc("/config/reset", g.resetHandler)

	return mux
}

func (g *GqlFakeRouter) verifyHandler(w http.ResponseWriter, r *http.Request) {
	g.m.Lock()
	defer g.m.Unlock()

	if len(g.ResponseConfig) != 0 {
		sb := strings.Builder{}
		for k, v := range g.ResponseConfig {
			sb.WriteString(fmt.Sprintf("%s %s was expected to be called %d more times\n", k.Type, k.Name, len(v)))
		}
		writeError(w, http.StatusInternalServerError, errors.New(sb.String()))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (g *GqlFakeRouter) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not POST"))
		return
	}

	body := ConfigRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	g.m.Lock()
	defer g.m.Unlock()

	g.ResponseConfig[body.GraphqlQueryKey] = append(g.ResponseConfig[body.GraphqlQueryKey], body.Response)
}

func (g *GqlFakeRouter) graphqlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not POST"))
		return
	}

	jsonBody := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&jsonBody)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	query, GQLerr := gqlparser.LoadQuery(g.Schema, jsonBody["query"].(string))
	if GQLerr != nil {
		writeError(w, http.StatusInternalServerError, GQLerr[0])
		return
	}

	queryField, ok := query.Operations[0].SelectionSet[0].(*ast.Field)
	if !ok {
		writeError(w, http.StatusInternalServerError, GQLerr[0])
		return
	}

	queryName := queryField.Name
	queryType := string(query.Operations[0].Operation)
	key := GraphqlQueryKey{
		Type: queryType,
		Name: queryName,
	}
	response, err := g.getResponseByKey(key)
	if err != nil {
		writeGQLError(w, err.Error())
		return
	}

	m, isMap := response.(map[string]interface{})
	if isMap {
		if errResponse, ok := m["error"]; ok {
			writeGQLError(w, errResponse.(string))
			return
		}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func (g *GqlFakeRouter) resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not POST"))
		return
	}

	g.m.Lock()
	defer g.m.Unlock()

	g.ResponseConfig = make(map[GraphqlQueryKey][]interface{})
}

func (g *GqlFakeRouter) getResponseByKey(key GraphqlQueryKey) (interface{}, error) {
	g.m.Lock()
	defer g.m.Unlock()
	if len(g.ResponseConfig[key]) == 0 {
		return nil, errors.New("no GQL response configured")
	}

	response := g.ResponseConfig[key][0]
	g.ResponseConfig[key] = g.ResponseConfig[key][1:]
	if len(g.ResponseConfig[key]) == 0 {
		delete(g.ResponseConfig, key)
	}
	return response, nil
}
