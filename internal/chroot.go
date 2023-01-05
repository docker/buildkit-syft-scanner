// Copyright 2022 buildkit-syft-scanner authors
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

package internal

import (
	"syscall"
)

// withChroot executes a target function inside a chroot environment.
//
// After the function is executed, the chroot is exited and the working
// directory is restored.
func withChroot(dir string, f func() error) error {
	// save previous state
	oldfd, err := syscall.Open("/", syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	oldwd, err := syscall.Getwd()
	if err != nil {
		return err
	}

	// set new state
	if err := syscall.Chroot(dir); err != nil {
		return err
	}
	if err := syscall.Chdir("/"); err != nil {
		return err
	}

	// execute target function
	err = f()

	// restore previous state
	if err2 := syscall.Fchdir(oldfd); err2 != nil && err == nil {
		return err2
	}
	if err2 := syscall.Chroot("."); err2 != nil && err == nil {
		return err2
	}
	if err2 := syscall.Chdir(oldwd); err2 != nil && err == nil {
		return err2
	}

	// cleanup
	if err2 := syscall.Close(oldfd); err2 != nil && err == nil {
		return err2
	}

	return err
}
