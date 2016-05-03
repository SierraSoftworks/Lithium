package license

import (
	"reflect"
	"testing"
	"time"
)

const demoData = `{"meta":{"id":"0","activates":"1970-01-01T00:00:00Z","expires":"1970-01-01T00:00:00Z","pack":{"test":{"count":0,"payload":{},"pack":{"test":{"count":1,"payload":{"x":1}}}}}},"payload":{"x":1}}`

var demoPayload = map[string]interface{}{
	"x": 1.0,
}

func TestEncodeData(t *testing.T) {
	testTime, err := time.Parse("01/02/2006 15:04:05", "01/01/1970 00:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data := Data{
		Meta: &Metadata{
			ID:          "0",
			ActivatesOn: testTime,
			ExpiresOn:   testTime,
			Pack: map[string]*Template{
				"test": &Template{
					Count:   0,
					Payload: map[string]interface{}{},
					Pack: map[string]*Template{
						"test": {
							Count: 1,
							Payload: map[string]interface{}{
								"x": 1,
							},
						},
					},
				},
			},
		},
		Payload: demoPayload,
	}

	b, err := EncodeData(&data)
	if err != nil {
		t.Fatal(err)
	}

	s := string(b)
	if s != demoData {
		t.Errorf("expected encoded data to be\nwant: %s\ngot:  %s", demoData, s)
	}
}

func TestDecodeData(t *testing.T) {
	testTime, err := time.Parse("01/02/2006 15:04:05", "01/01/1970 00:00:00")
	if err != nil {
		t.Fatal(err)
	}

	data, err := DecodeData([]byte(demoData))
	if err != nil {
		t.Fatal(err)
	}

	if data == nil {
		t.Fatal("expected returned data to be non-nil")
	}

	if data.Meta == nil {
		t.Fatal("expected data.Meta to be non-nil")
	}

	if data.Meta.ID != "0" {
		t.Errorf("expected ID to be '0', got '%v'", data.Meta.ID)
	}

	if data.Meta.ActivatesOn != testTime {
		t.Errorf("expected activate time to be '%v' but got '%v'", testTime, data.Meta.ActivatesOn)
	}

	if data.Meta.ExpiresOn != testTime {
		t.Errorf("expected expiry time to be '%v' but got '%v'", testTime, data.Meta.ExpiresOn)
	}

	if data.Meta.Pack == nil {
		t.Fatal("expected license pack to be non-nil")
	}

	testPack, exists := data.Meta.Pack["test"]
	if !exists {
		t.Error("expected license pack to include a test type")
	} else {
		if testPack.Count != 0 {
			t.Errorf("expected a count of 0, got %v", testPack.Count)
		}

		if testPack.Payload == nil {
			t.Error("expected payload to be non-nil")
		}

		if !reflect.DeepEqual(testPack.Payload, map[string]interface{}{}) {
			t.Error("expected payload to be an empty object")
		}

		if testPack.Pack == nil {
			t.Fatal("expected pack to contain a sublicense pack")
		}

		testPack2, exists := testPack.Pack["test"]
		if !exists {
			t.Error("expected a 'test' sublicense to exist")
		} else {
			if testPack2.Count != 1 {
				t.Errorf("expected sublicense to have a count of 1, got %d", testPack2.Count)
			}

			if testPack2.Payload == nil {
				t.Error("expected sublicense to have a payload")
			} else {
				if testPack2.Payload["x"] != 1.0 {
					t.Errorf("Expected sublicense payload.x to be 1, got %#v", testPack2.Payload["x"])
				}
			}

			if testPack2.Pack != nil {
				t.Error("expected sublicense to not have a license pack")
			}
		}
	}

	if data.Payload == nil {
		t.Fatal("expected data.Payload to be non-nil")
	}

	if !reflect.DeepEqual(data.Payload, demoPayload) {
		t.Errorf("expected data.Payload to be { x: 1 }, got %v", data.Payload)
	}

}
