package server

import (
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/vrischmann/envconfig"
)

func Token() string {
	tokenConfig := Config{}
	err := envconfig.InitWithPrefix(&tokenConfig, "APP")
	if err != nil {
		log.Fatal(err)
	}

	var dexToken string

	log.Info("Get Dex id_token")
	if tokenConfig.IsWithToken {
		tokenConfig.Log = log.Infof
		ts := newTokenServer(&tokenConfig)
		dexToken = waitForToken(ts)
	} else {
		dexToken, err = idtokenprovider.GetDexToken()
		if err != nil {
			log.Fatal(errors.Wrap(err, "while getting dex token"))
		}
	}

	return dexToken
}
