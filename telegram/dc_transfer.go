package telegram

import (
	"context"

	"github.com/owpengram/owpengram-td/telegram/auth"
	"github.com/owpengram/owpengram-td/telegram/internal/manager"
	"github.com/owpengram/owpengram-td/tg"
)

func (c *Client) dcTransferSetup(dcID int) manager.SetupCallback {
	return func(ctx context.Context, invoker tg.Invoker) error {
		// Run export/import authorization only when the connection is already up.
		_, err := c.transfer(ctx, tg.NewClient(invoker), dcID)
		if auth.IsUnauthorized(err) {
			return nil
		}
		return err
	}
}
