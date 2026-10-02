package txmanagertest

import (
	"context"

	"gorm.io/gorm"
)

func (f *Fake) GetConnection(ctx context.Context) *gorm.DB {
	return f.GetRawConnection(ctx)
}
