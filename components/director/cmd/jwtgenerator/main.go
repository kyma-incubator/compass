package main

import (
	"flag"
	"fmt"

	"github.com/form3tech-oss/jwt-go"
)

type Claims struct {
	Scopes string `json:"scopes"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

func main() {
	tenantFlag := flag.String("tenant", "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae", "the tenant ID that should be in the token claims")
	flag.Parse()

	claims := Claims{
		Scopes:         "application:read automatic_scenario_assignment:write automatic_scenario_assignment:read health_checks:read application:write runtime:write label_definition:write label_definition:read runtime:read tenant:read",
		Tenant:         *tenantFlag,
		StandardClaims: jwt.StandardClaims{},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		panic(err)
	}

	fmt.Println(signedToken)
}
