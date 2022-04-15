package authenticator

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/lestrrat-go/jwx/jwk"
)

/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// JWTKeyIterator missing godoc
type JWTKeyIterator struct {
	AlgorithmCriteria func(string) bool
	IDCriteria        func(string) bool
	ResultingKey      interface{}
	AllKeys           []interface{}
}

// Visit missing godoc
func (keyIterator *JWTKeyIterator) Visit(_ int, value interface{}) error {
	key, ok := value.(jwk.Key)
	if !ok {
		return apperrors.NewInternalError("unable to parse key")
	}

	keyIterator.AllKeys = append(keyIterator.AllKeys, key)

	if keyIterator.AlgorithmCriteria(key.Algorithm()) && keyIterator.IDCriteria(key.KeyID()) {
		var rawKey interface{}
		if err := key.Raw(&rawKey); err != nil {
			return err
		}

		keyIterator.ResultingKey = rawKey
	}

	return nil
}
