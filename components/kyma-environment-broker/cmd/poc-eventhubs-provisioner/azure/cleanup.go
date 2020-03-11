package azure

import (
	"context"
	"log"
)

func Cleanup(ctx context.Context, config *Config, groupName string) error {
	if config.keepResources {
		log.Println("Resource cleanup is disabled")
		return nil
	}

	if _, err := DeleteGroup(ctx, config, groupName); err != nil {
		log.Printf("Failed to delete Azure Resource Group [%s]", groupName)
		return err
	}
	log.Printf("Deleted Azure Resource Group [%s]", groupName)

	return nil
}
