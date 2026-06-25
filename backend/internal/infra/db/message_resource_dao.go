package db

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/Leros/backend/types"
)

// BatchCreateMessageResources creates multiple message_resource records in a single transaction.
func BatchCreateMessageResources(ctx context.Context, db *gorm.DB, records []*types.MessageResource) error {
	if len(records) == 0 {
		return nil
	}

	return db.WithContext(ctx).Create(records).Error
}
