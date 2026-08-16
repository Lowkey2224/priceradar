package dto

import (
	"reflect"
	"strings"
	"testing"
)

func TestDeserialization(t *testing.T) {
	jsonString := `{"title":"Product Title", "urls":["http://foo.bar", "http://example.com"], "targetPrice":123}`
	expected := CreateProductRequest{
		Title:       "Product Title",
		Urls:        []string{"http://foo.bar", "http://example.com"},
		TargetPrice: 123,
	}
	actual := CreateProductRequest{}
	err := actual.New(strings.NewReader(jsonString))
	if err != nil {
		t.Fatalf("New Failed with error %v", err)
	}
	if actual.TargetPrice != expected.TargetPrice {
		t.Errorf("New() TargetPrice = %v, want %v", actual.TargetPrice, expected.TargetPrice)
	}
	if actual.Title != expected.Title {
		t.Errorf("New() Title = %v, want %v", actual.Title, expected.Title)
	}
	if !reflect.DeepEqual(actual.Urls, expected.Urls) {
		t.Errorf("New() Urls = %v, want %v", actual.Urls, expected.Urls)
	}
	if err = actual.Valid(); err != nil {
		t.Errorf("Valid() expected to be nil, got %v", err)
	}
}

func TestDeserializationFailsGracefully(t *testing.T) {

	actual := CreateProductRequest{}
	err := actual.New(strings.NewReader(`{"targetPrice":123.12}`))
	expected := "json: cannot unmarshal number 123.12 into Go struct field CreateProductRequest.targetPrice of type int64"
	if err.Error() != expected {
		t.Fatalf("New() got error `%v`, want `%v`", expected, err)
	}
	err = actual.New(strings.NewReader(``))
	expected = "EOF"
	if err.Error() != expected {
		t.Fatalf("New() got error `%v`, want `%v`", expected, err)
	}
}

func TestValid(t *testing.T) {
	jsonString := `{}`
	actual := CreateProductRequest{}
	err := actual.New(strings.NewReader(jsonString))

	if err != nil {
		t.Fatalf("New() got error `%v`", err)
	}
	if err = actual.Valid(); err == nil {
		t.Fatalf("Valid() expected error, got nil")
	}
	expected := "Urls must contain Elements,targetPrice must be bigger than 0,title must not be empty"
	if err.Error() != expected {
		t.Fatalf("Valid() got error `%v`, want `%v`", expected, err)
	}
}
