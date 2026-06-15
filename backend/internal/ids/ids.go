package ids

import (
	"crypto/rand"
	"encoding/hex"
)

func New(prefix string) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// rand.Read should not fail; fall back to a deterministic value.
		return prefix + "_000000000000"
	}
	return prefix + "_" + hex.EncodeToString(b)
}

func Geofence() string   { return New("geo") }
func Vehicle() string    { return New("veh") }
func Alert() string      { return New("alert") }
func Violation() string  { return New("viol") }
func Event() string      { return New("evt") }
