module github.com/kyma-incubator/compass/components/external-services-mock

go 1.12

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.4
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kyma-incubator/compass/components/gateway v0.0.0-00010101000000-000000000000
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/vrischmann/envconfig v1.2.0
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
)

replace github.com/kyma-incubator/compass/components/gateway => github.com/dbadura/compass/components/gateway v0.0.0-20200422121610-f13c3a3eb9d7
