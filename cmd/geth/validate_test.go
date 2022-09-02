package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"
)

func TestNewValidateInfo(t *testing.T) {

	key, _ := crypto.GenerateKey()

	v, err := NewValidateInfo(key)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(v.ToJson())
	fmt.Println(v.Validate())

}

func BenchmarkValidate(b *testing.B) {
	key, _ := crypto.GenerateKey()

	v, err := NewValidateInfo(key)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		v.Validate()
	}
}

func BenchmarkNewValidateInfo(b *testing.B) {
	key, _ := crypto.GenerateKey()
	for i := 0; i < b.N; i++ {
		NewValidateInfo(key)
	}
}
