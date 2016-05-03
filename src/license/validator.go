package license

import (
	"errors"
	"time"
)

// IsValid determines whether a license object is valid for use.
// It does so by checking that the metadata is present, that it
// is within the validity period and that there is a defined payload.
func (l *Data) IsValid() (bool, error) {
	if l.Meta == nil {
		return false, errors.New("license metadata not defined")
	}

	now := time.Now()

	if now.Before(l.Meta.ActivatesOn) {
		return false, errors.New("license has not yet activated due to time constraint")
	}

	if now.After(l.Meta.ExpiresOn) {
		return false, errors.New("license has expired due to time constraint")
	}

	if l.Payload == nil {
		return false, errors.New("license payload was not defined")
	}

	return true, nil
}
