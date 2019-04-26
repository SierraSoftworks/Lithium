package license

import (
	"testing"
	"time"
)

func TestDataIsValid(t *testing.T) {
	l := Data{}

	v, err := l.IsValid()
	if v {
		t.Error("expected empty license to be invalid")
	}

	if err == nil {
		t.Error("expected error message to be present")
	} else if err.Error() != "license metadata not defined" {
		t.Errorf("expected error message to be descriptive, got '%s'", err)
	}

	l.Meta = &Metadata{
		ActivatesOn: time.Now().Add(5 * time.Second),
	}

	v, err = l.IsValid()
	if v {
		t.Error("expected time delayed license to not be valid")
	}

	if err == nil {
		t.Error("expected error message to be present")
	} else if err.Error() != "license has not yet activated due to time constraint" {
		t.Errorf("expected error message to be descriptive, got '%s'", err)
	}

	l.Meta = &Metadata{
		ActivatesOn: time.Now().Add(-15 * time.Second),
		ExpiresOn:   time.Now().Add(-5 * time.Second),
	}

	v, err = l.IsValid()
	if v {
		t.Error("expected expired license to not be valid")
	}

	if err == nil {
		t.Error("expected error message to be present")
	} else if err.Error() != "license has expired due to time constraint" {
		t.Errorf("expected error message to be descriptive, got '%s'", err)
	}

	l.Meta = &Metadata{
		ActivatesOn: time.Now().Add(-5 * time.Second),
		ExpiresOn:   time.Now().Add(5 * time.Second),
	}

	v, err = l.IsValid()
	if v {
		t.Error("expected license with no payload to be invalid")
	}

	if err == nil {
		t.Error("expected error message to be present")
	} else if err.Error() != "license payload was not defined" {
		t.Errorf("expected error message to be descriptive, got '%s'", err)
	}

	l.Payload = map[string]interface{}{}
	v, err = l.IsValid()
	if !v {
		t.Error("expected license with good metadata and payload to be valid")
	}

	if err != nil {
		t.Errorf("expected valid license to have no error, got '%s'", err.Error())
	}
}
