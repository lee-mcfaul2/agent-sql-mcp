package auth

import "encoding/base64"

// unwrapDexSub returns the inner user_id when sub is in Dex's
// password-connector wrapped form, otherwise returns sub unchanged.
//
// Dex's password connector emits `sub` as a base64-encoded protobuf
// wrapping the configured userID + connector ID, regardless of what's
// configured. For demo userID "00000000-0000-0000-0000-000000000001"
// and connector "local" the wire form is:
//
//	field 1 (LEN, tag=0x0a): 36 bytes "00000000-..."
//	field 2 (LEN, tag=0x12):  5 bytes "local"
//
// base64-encoded -> "CiQwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDESBWxvY2Fs".
//
// demoPermsBySub is keyed by the BARE UUID; without unwrapping every
// password-connector caller misses the map and is denied. We decode in
// a deliberately strict way: must base64-decode AND start with 0x0a
// (proto field 1, wire type 2) AND the next byte must be a plausible
// length (1..255). Anything else returns the input unchanged so a real
// IdP that happens to emit a base64-looking sub is unaffected.
func unwrapDexSub(sub string) string {
	for _, enc := range []*base64.Encoding{base64.RawStdEncoding, base64.RawURLEncoding} {
		raw, err := enc.DecodeString(sub)
		if err != nil {
			continue
		}
		if len(raw) < 2 || raw[0] != 0x0a {
			continue
		}
		uidLen := int(raw[1])
		if uidLen == 0 || uidLen > len(raw)-2 {
			continue
		}
		return string(raw[2 : 2+uidLen])
	}
	return sub
}
