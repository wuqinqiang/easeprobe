/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
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

package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/megaease/easeprobe/global"
	"github.com/megaease/easeprobe/probe"
	"github.com/megaease/easeprobe/probe/base"
	log "github.com/sirupsen/logrus"
)

// Shell implements a config for shell command (os.Exec)
type Shell struct {
	base.DefaultOptions `yaml:",inline"`
	Command             string   `yaml:"cmd"`
	Args                []string `yaml:"args,omitempty"`
	Env                 []string `yaml:"env,omitempty"`
	Contain             string   `yaml:"contain,omitempty"`
	NotContain          string   `yaml:"not_contain,omitempty"`
}

// Config Shell Config Object
func (s *Shell) Config(gConf global.ProbeSettings) error {
	kind := "shell"
	tag := ""
	name := s.ProbeName
	s.DefaultOptions.Config(gConf, kind, tag, name,
		probe.CommandLine(s.Command, s.Args), s.DoProbe)

	log.Debugf("[%s] configuration: %+v, %+v", s.ProbeKind, s, s.Result())
	return nil
}

// DoProbe return the checking result
func (s *Shell) DoProbe() (bool, string) {

	ctx, cancel := context.WithTimeout(context.Background(), s.ProbeTimeout)
	defer cancel()

	for _, e := range s.Env {
		v := strings.Split(e, "=")
		os.Setenv(v[0], v[1])
	}

	cmd := exec.CommandContext(ctx, s.Command, s.Args...)
	output, err := cmd.CombinedOutput()

	status := true
	message := "Shell Command has been Run Successfully!"

	if err != nil {
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		message = fmt.Sprintf("Error: %v, ExitCode(%d), Output:%s",
			err, exitCode, probe.CheckEmpty(string(output)))
		log.Errorf(message)
		status = false
	}
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CommandLine(s.Command, s.Args))
	log.Debugf("[%s / %s] - %s", s.ProbeKind, s.ProbeName, probe.CheckEmpty(string(output)))

	if err := probe.CheckOutput(s.Contain, s.NotContain, string(output)); err != nil {
		log.Errorf("[%s / %s] - %v", s.ProbeKind, s.ProbeName, err)
		s.ProbeResult.Message = fmt.Sprintf("Error: %v", err)
		status = false
	}

	return status, message
}
