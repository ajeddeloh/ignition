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

package sgdisk

import (
	"fmt"
	"os/exec"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/log"
)

const sgdiskPath = "/sbin/sgdisk"

type Operation struct {
	logger    *log.Logger
	dev       string
	wipe      bool
	parts     []types.Partition
	deletions []int
}

// Begin begins an sgdisk operation
func Begin(logger *log.Logger, dev string) *Operation {
	return &Operation{logger: logger, dev: dev}
}

// CreatePartition adds the supplied partition to the list of partitions to be created as part of an operation.
func (op *Operation) CreatePartition(p types.Partition) {
	op.parts = append(op.parts, p)
}

func (op *Operation) DeletePartition(num int) {
	op.deletions = append(op.deletions, num)
}

// WipeTable toggles if the table is to be wiped first when commiting this operation.
func (op *Operation) WipeTable(wipe bool) {
	op.wipe = wipe
}

// Commit commits an partitioning operation.
func (op *Operation) Commit() error {
	if op.wipe {
		cmd := exec.Command(sgdiskPath, "--zap-all", op.dev)
		if _, err := op.logger.LogCmd(cmd, "wiping table on %q", op.dev); err != nil {
			op.logger.Info("potential error encountered while wiping table... retrying")
			cmd = exec.Command(sgdiskPath, "--zap-all", op.dev)
			if _, err := op.logger.LogCmd(cmd, "wiping table on %q", op.dev); err != nil {
				return fmt.Errorf("wipe failed: %v", err)
			}
		}
	}

	opts := []string{}

	// Do all deletions before creations
	for _, partition := range op.deletions {
		opts = append(opts, fmt.Sprintf("--delete=%d", partition))
	}

	for _, p := range op.parts {
		opts = append(opts, fmt.Sprintf("--new=%d:%d:+%d", p.Number, p.Start, p.Size))
		if p.Label != nil {
			opts = append(opts, fmt.Sprintf("--change-name=%d:%s", p.Number, *p.Label))
		}
		if p.TypeGUID != "" {
			opts = append(opts, fmt.Sprintf("--typecode=%d:%s", p.Number, p.TypeGUID))
		}
		if p.GUID != "" {
			opts = append(opts, fmt.Sprintf("--partition-guid=%d:%s", p.Number, p.GUID))
		}
	}

	if len(opts) == 0 {
		return nil
	}

	opts = append(opts, op.dev)
	cmd := exec.Command(sgdiskPath, opts...)
	if _, err := op.logger.LogCmd(cmd, "creating %d partitions on %q", len(op.parts), op.dev); err != nil {
		return fmt.Errorf("create partitions failed: %v", err)
	}

	return nil
}
