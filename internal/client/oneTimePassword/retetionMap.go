package oneTimePassword

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RetentionMap map[string]OTP

func NewRetentionMap(ctx context.Context, period time.Duration) RetentionMap {
	rm := make(RetentionMap)
	go rm.Retention(ctx, period)
	return rm
}

func (r RetentionMap) NewOTP() OTP {
	o := OTP{
		OneTimePassword: uuid.NewString(),
		Created:         time.Now(),
	}

	r[o.OneTimePassword] = o
	return o
}

func (r RetentionMap) Verify(otp string) bool {
	if _, ok := r[otp]; ok {
		delete(r, otp)
		return true
	}
	return false
}

func (r RetentionMap) Retention(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(400 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			for _, o := range r {
				if o.Created.Add(period).Before(time.Now()) {
					delete(r, o.OneTimePassword)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
