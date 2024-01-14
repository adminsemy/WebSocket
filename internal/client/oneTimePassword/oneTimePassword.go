package oneTimePassword

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	oneTimePassword string
	Created         time.Time
}

func NewOTP(rm RetentionMap) OTP {
	o := OTP{
		oneTimePassword: uuid.NewString(),
		Created:         time.Now(),
	}

	rm[o.oneTimePassword] = o
	return o
}
