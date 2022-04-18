package app

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

const SALT = "_salt"

type PasswordEncoder interface {
	Encode(password string) string
}

type encoder struct {
	sha512 hash.Hash
}

func (e encoder) Encode(password string) string {
	return e.getHash(password + SALT)
}

func (e encoder) getHash(text string) string {
	result := sha512.Sum512_256([]byte(text))
	return hex.EncodeToString(result[:])
}

func NewPasswordEncoder() PasswordEncoder {
	return &encoder{
		sha512: sha512.New(),
	}
}
