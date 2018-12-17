// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v3_0_experimental

import (
	"github.com/coreos/ignition/config/shared/errors"
	"github.com/coreos/ignition/config/util"
	"github.com/coreos/ignition/config/v3_0_experimental/types"
	"github.com/coreos/ignition/config/validate"
	"github.com/coreos/ignition/config/validate/report"

	json "github.com/ajeddeloh/go-json"
	"github.com/coreos/go-semver/semver"
)

// Parse parses the raw config into a types.Config struct and generates a report of any
// errors, warnings, info, and deprecations it encountered
func Parse(rawConfig []byte) (types.Config, report.Report, error) {
	if isEmpty(rawConfig) {
		return types.Config{}, report.Report{}, errors.ErrEmpty
	} else if isCloudConfig(rawConfig) {
		return types.Config{}, report.Report{}, errors.ErrCloudConfig
	}

	var config types.Config

	if err := json.Unmarshal(rawConfig, &config); err != nil {
		rpt, err := util.HandleParseErrors(rawConfig)
		// HandleParseErrors always returns an error
		return types.Config{}, rpt, err
	}

	version, err := semver.NewVersion(config.Ignition.Version)

	if err != nil || *version != types.MaxVersion {
		return types.Config{}, report.Report{}, errors.ErrUnknownVersion
	}

	rpt := validate.ValidateConfig(rawConfig, config)
	if rpt.IsFatal() {
		return types.Config{}, rpt, errors.ErrInvalid
	}

	return config, rpt, nil
}

func isEmpty(userdata []byte) bool {
	return len(userdata) == 0
}
