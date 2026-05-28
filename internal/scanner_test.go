// Copyright 2024 buildkit-syft-scanner authors
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
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestWriteStatement_ClosesFileOnEncodeError verifies that writeStatement does
// not leak the file descriptor when the destination directory is not writable,
// and that it succeeds and produces valid JSON when the destination is writable.
func TestWriteStatement_ClosesFileOnEncodeError(t *testing.T) {
	t.Run("returns error when destination is not writable", func(t *testing.T) {
		// Create a temp dir and make it read-only so os.Create fails.
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o555); err != nil {
			t.Skipf("cannot chmod temp dir (may be running as root): %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

		outputPath := filepath.Join(dir, "out.spdx.json")
		err := writeStatement(outputPath, map[string]string{"key": "value"})
		if err == nil {
			t.Fatal("expected an error writing to read-only directory, got nil")
		}
	})

	t.Run("writes valid JSON and closes file on success", func(t *testing.T) {
		dir := t.TempDir()
		outputPath := filepath.Join(dir, "out.spdx.json")

		payload := map[string]string{"predicateType": "https://spdx.dev/Document"}
		if err := writeStatement(outputPath, payload); err != nil {
			t.Fatalf("writeStatement returned unexpected error: %v", err)
		}

		data, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("output file not readable after writeStatement: %v", err)
		}
		var got map[string]string
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("output file is not valid JSON: %v", err)
		}
		if got["predicateType"] != payload["predicateType"] {
			t.Errorf("expected predicateType %q, got %q", payload["predicateType"], got["predicateType"])
		}
	})
}

// TestLoadPathFromEnvironment_RequiredMissing verifies that a missing required
// environment variable produces an error rather than an empty path.
func TestLoadPathFromEnvironment_RequiredMissing(t *testing.T) {
	const key = "BUILDKIT_TEST_MISSING_VAR_XYZ"
	t.Setenv(key, "")
	os.Unsetenv(key)

	_, err := loadPathFromEnvironment(key, true)
	if err == nil {
		t.Fatal("expected error for missing required variable, got nil")
	}
}

// TestLoadPathFromEnvironment_OptionalMissing verifies that a missing optional
// variable returns an empty string with no error.
func TestLoadPathFromEnvironment_OptionalMissing(t *testing.T) {
	const key = "BUILDKIT_TEST_OPTIONAL_VAR_XYZ"
	os.Unsetenv(key)

	got, err := loadPathFromEnvironment(key, false)
	if err != nil {
		t.Fatalf("expected no error for missing optional variable, got: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string for missing optional variable, got %q", got)
	}
}

// TestLoadPathFromEnvironment_PathNotExist verifies that a variable pointing to
// a non-existent path returns an error.
func TestLoadPathFromEnvironment_PathNotExist(t *testing.T) {
	const key = "BUILDKIT_TEST_NONEXIST_VAR_XYZ"
	t.Setenv(key, "/nonexistent/path/that/cannot/exist")

	_, err := loadPathFromEnvironment(key, true)
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}
