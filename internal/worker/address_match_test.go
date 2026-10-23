package worker

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestMayMatchAddress(t *testing.T) {
	address := common.HexToAddress("0x0123456789abcdef0123456789abcdef01234567")
	hex := address.Hex()
	for _, tc := range []struct {
		name, prefix, suffix string
		want                 bool
	}{
		{"empty", "", "", true},
		{"fixed zero", "0", "", true},
		{"fixed prefix", "0x", "", true},
		{"odd prefix nibble", "0x0", "", true},
		{"even prefix nibbles", "0x01", "", true},
		{"odd suffix nibbles", "", "567", true},
		{"even suffix nibbles", "", "4567", true},
		{"both ends", "0x012", "67", true},
		{"letter case deferred", "0x0123456789ABCDEF", "", true},
		{"suffix crosses x", "", hex[1:], true},
		{"full suffix", "", hex, true},
		{"full prefix", hex, "", true},
		{"wrong prefix nibble", "0x02", "", false},
		{"wrong suffix nibble", "", "68", false},
		{"fixed x is case sensitive", "0X", "", false},
		{"suffix fixed x is case sensitive", "", "X" + hex[2:], false},
		{"invalid prefix character", "0xg", "", false},
		{"invalid suffix character", "", "!", false},
		{"non-ASCII character", "0xé", "", false},
		{"long prefix", hex + "0", "", false},
		{"long suffix", "", "0" + hex, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := mayMatchAddress(address, tc.prefix, tc.suffix); got != tc.want {
				t.Fatalf("mayMatchAddress(%q, %q) = %v, want %v", tc.prefix, tc.suffix, got, tc.want)
			}
		})
	}

	// The raw filter must retain case-only candidates for the final EIP-55 check.
	checksumAddress := common.HexToAddress("0x52908400098527886E0F7030069857D2E4169EE7")
	for _, rule := range []struct{ prefix, suffix string }{
		{prefix: "0x52908400098527886e"},
		{suffix: "ee7"},
	} {
		if !mayMatchAddress(checksumAddress, rule.prefix, rule.suffix) {
			t.Fatal("raw filter rejected a case-only candidate")
		}
		if strings.HasPrefix(checksumAddress.Hex(), rule.prefix) && strings.HasSuffix(checksumAddress.Hex(), rule.suffix) {
			t.Fatal("case-only candidate unexpectedly passed the exact checksum comparison")
		}
	}

	// Every exact match, including rules that overlap, must survive the filter.
	for seed := 0; seed < 8; seed++ {
		var candidate common.Address
		for i := range candidate {
			candidate[i] = byte(seed*37 + i*19)
		}
		exact := candidate.Hex()
		for prefixLength := 0; prefixLength <= len(exact); prefixLength++ {
			for suffixLength := 0; suffixLength <= len(exact); suffixLength++ {
				prefix, suffix := exact[:prefixLength], exact[len(exact)-suffixLength:]
				if !mayMatchAddress(candidate, prefix, suffix) {
					t.Fatalf("rejected exact match for %s: prefix %q, suffix %q", exact, prefix, suffix)
				}
			}
		}
	}
}
