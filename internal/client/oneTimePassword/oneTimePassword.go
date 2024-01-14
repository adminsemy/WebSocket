package oneTimePassword

import (
	"time"
)

type OTP struct {
	OneTimePassword string
	Created         time.Time
}
