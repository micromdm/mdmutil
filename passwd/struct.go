// DO NOT EDIT
// generated from https://github.com/apple/device-management:67045e2fa06f528b196c01edee6a8bf88b844beb/other/passwordhash.yaml

package passwd

// A dictionary that contains the password hash for the account.
type PasswordHash struct {
	// A dictionary that contains the `entropy`, `iterations`, and `salt` elements to create the password hash using the CommonCrypto libraries, or equivalent. Convert this dictionary to binary data before setting it as the value for the password hash.
	SALTEDSHA512PBKDF2 SALTEDSHA512PBKDF2 `plist:"SALTED-SHA512-PBKDF2"`
}

// A dictionary that contains the `entropy`, `iterations`, and `salt` elements to create the password hash using the CommonCrypto libraries, or equivalent. Convert this dictionary to binary data before setting it as the value for the password hash.
type SALTEDSHA512PBKDF2 struct {
	// The derived key from the password hash; for example, from `CCKeyDerivationPBKDF()`.
	Entropy []byte `plist:"entropy"`
	// The number of iterations; for example, from `CCCalibratePBKDF()` using a minimum hash time of 100 milliseconds, or if unknown, a number in the range of 20,000 to 40,000 iterations.
	Iterations int64 `plist:"iterations"`
	// The 32-byte randomized data; for example, from `CCRandomCopyBytes()`.
	Salt []byte `plist:"salt"`
}
