// Copyright 2019 Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package template

import (
	"fmt"
	"io/ioutil"

	tmv1beta1 "github.com/gardener/test-infra/pkg/apis/testmachinery/v1beta1"
	"github.com/gardener/test-infra/pkg/testrunner/componentdescriptor"
	"github.com/gardener/test-infra/pkg/testrunner/result"
	"github.com/gardener/test-infra/pkg/util"
	log "github.com/sirupsen/logrus"
)

// Render renders a helm chart with containing testruns, adds the provided parameters and values, and returns the parsed and modified testruns.
func Render(tmKubeconfigPath string, parameters *TestrunParameters, metadata *result.Metadata) ([]*tmv1beta1.Testrun, error) {
	versions, err := getK8sVersions(parameters)
	if err != nil {
		log.Fatal(err.Error())
	}

	if parameters.ComponentDescriptorPath != "" {
		data, err := ioutil.ReadFile(parameters.ComponentDescriptorPath)
		if err != nil {
			log.Warnf("Cannot read component descriptor file %s: %s", parameters.ComponentDescriptorPath, err.Error())
		}
		components, err := componentdescriptor.GetComponents(data)
		if err != nil {
			log.Warnf("Cannot decode and parse BOM %s", err.Error())
		} else {
			metadata.BOM = components
		}
	}

	files, err := RenderChart(tmKubeconfigPath, parameters, versions)
	if err != nil {
		return nil, err
	}

	// parse the rendered testruns and add locations from BOM of a bom was provided.
	testruns := []*tmv1beta1.Testrun{}
	for _, fileContent := range files {
		tr, err := util.ParseTestrun([]byte(fileContent))
		if err != nil {
			log.Warnf("Cannot parse rendered file: %s", err.Error())
		}

		// Add current dependency repositories to the testrun location.
		// This gives us all dependent repositories as well as there deployed version.
		addBOMLocationsToTestrun(&tr, metadata.BOM)

		testruns = append(testruns, &tr)
	}

	if len(testruns) == 0 {
		return nil, fmt.Errorf("No testruns in the helm chart at %s", parameters.TestrunChartPath)
	}

	return testruns, nil
}