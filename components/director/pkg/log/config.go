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

package log

type Config struct {
	Level                  string `envconfig:"LOG_LEVEL,default=info"`
	Format                 string `envconfig:"LOG_FORMAT,default=text"`
	Output                 string `envconfig:"LOG_OUTPUT,default=/dev/stdout"`
	BootstrapCorrelationID string `envconfig:"LOG_BOOTSTRAP_CORRELATION_ID,default=bootstrap_correlation_id"`
}
