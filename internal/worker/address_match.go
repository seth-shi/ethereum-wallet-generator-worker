package worker

import "github.com/ethereum/go-ethereum/common"

// mayMatchAddress rejects candidates before constructing their EIP-55 string.
// Letter case is checked against the checksummed address after this filter.
func mayMatchAddress(address common.Address, prefix, suffix string) bool {
	const addressLength = 2 + 2*common.AddressLength
	if len(prefix) > addressLength || len(suffix) > addressLength {
		return false
	}
	for i := range prefix {
		if !matchesAddressCharacter(address, i, prefix[i]) {
			return false
		}
	}
	for i := range suffix {
		if !matchesAddressCharacter(address, addressLength-len(suffix)+i, suffix[i]) {
			return false
		}
	}
	return true
}

func matchesAddressCharacter(address common.Address, position int, character byte) bool {
	if position < 2 {
		return character == "0x"[position]
	}
	if character >= 'A' && character <= 'F' {
		character += 'a' - 'A'
	}
	position -= 2
	nibble := address[position/2]
	if position%2 == 0 {
		nibble >>= 4
	} else {
		nibble &= 0x0f
	}
	return character == "0123456789abcdef"[nibble]
}
