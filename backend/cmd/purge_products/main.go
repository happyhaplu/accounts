package main

import (
	"fmt"
	"log"

	"outcraftly/accounts/config"
	"outcraftly/accounts/database"
	"outcraftly/accounts/models"
)

func main() {
	cfg := config.Load()
	database.Connect(cfg)

	// Keep this utility explicit: remove product-linked access state first.
	subRes := database.DB.Where("1 = 1").Delete(&models.Subscription{})
	if subRes.Error != nil {
		log.Fatalf("failed to delete subscriptions: %v", subRes.Error)
	}

	prodRes := database.DB.Where("1 = 1").Delete(&models.Product{})
	if prodRes.Error != nil {
		log.Fatalf("failed to delete products: %v", prodRes.Error)
	}

	fmt.Printf("purge complete: deleted %d subscriptions, %d products\n", subRes.RowsAffected, prodRes.RowsAffected)
}
