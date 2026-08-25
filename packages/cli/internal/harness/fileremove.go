package harness

import (
	"fmt"
	"os"
)

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("harness: remove %s: %w", path, err)
	}

	return nil
}
