package models

import (
	"crypto/rand"
	"encoding/hex"
)

func NewEventID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

