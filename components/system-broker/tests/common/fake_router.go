package common

import (
	"encoding/json"
	"errors"
	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"io/ioutil"
	"net/http"
	"sync"
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
	m sync.RWMutex
	ResponseConfig map[GraphqlQueryKey]interface{}
	Schema *ast.Schema
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

	return &GqlFakeRouter{Schema: schema}, nil
}

func (g *GqlFakeRouter) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/config", g.configHandler)
	mux.HandleFunc("/graphql", g.graphqlHandler)

	return mux
}

func (g *GqlFakeRouter) configHandler(w http.ResponseWriter, r *http.Request)  {
	body := ConfigRequestBody{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}
	g.m.Lock()
	defer g.m.Unlock()

	g.ResponseConfig[body.GraphqlQueryKey] = body.Response
}

func (g *GqlFakeRouter) graphqlHandler(w http.ResponseWriter, r *http.Request)  {

	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}

	query , err := gqlparser.LoadQuery(g.Schema, string(requestBody))
	if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}
	queryField := query.Operations[0].SelectionSet[0].(*ast.Field)

	queryName := queryField.Name
	queryType := string(query.Operations[0].Operation)
	key := GraphqlQueryKey{
		Type: queryType,
		Name: queryName,
	}

	if err := json.NewEncoder(w).Encode(g.ResponseConfig[key]); err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}
}


