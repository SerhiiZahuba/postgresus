package files_utils

import (
	"fmt"
	"os"
)

func EnsureDirectories(directories []string) error {
	// Standard permissions for directories: owner
	// can read/write/execute, others can read/execute
	const directoryPermissions = 0755

	for _, directory := range directories {
		if err := os.MkdirAll(directory, directoryPermissions); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", directory, err)
		}
	}

	return nil
}
