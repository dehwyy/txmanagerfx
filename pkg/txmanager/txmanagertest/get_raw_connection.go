package txmanagertest

import (
	"context"

	"gorm.io/gorm"
)

func (f *Fake) GetRawConnection(ctx context.Context) *gorm.DB {
	if f.DB == nil {
		return nil
	}

	return f.DB.WithContext(ctx)
}
