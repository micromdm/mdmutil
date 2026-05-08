package passwd

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"math/big"
)

const (
	IterationsMin = 20000
	IterationsMax = 40000
)

// HashPassword derives a salted SHA-512 PBKDF2 key from plaintext.
// The struct returned is ready to be marshalled into an Apple Property List.
// Ostensibly for the `AccountConfiguration` and/or `SetAutoAdminPassword` MDM commands.
func HashPassword(randReader io.Reader, plaintext string) (*PasswordHash, error) {
	// iterations
	n, err := rand.Int(randReader, big.NewInt(IterationsMax-IterationsMin))
	if err != nil {
		return nil, fmt.Errorf("generate random iterations: %w", err)
	}
	iter := int(n.Int64()) + IterationsMin

	// salt
	salt := make([]byte, 32)
	_, err = io.ReadFull(randReader, salt)
	if err != nil {
		return nil, fmt.Errorf("error reading random: %w", err)
	}

	// key/hash
	key, err := pbkdf2.Key(sha512.New, plaintext, salt, iter, sha512.Size)
	if err != nil {
		return nil, fmt.Errorf("deriving PBKDF2 key for password hash: %w", err)
	}

	return &PasswordHash{
		SALTEDSHA512PBKDF2: SALTEDSHA512PBKDF2{
			Entropy:    key,
			Iterations: int64(iter),
			Salt:       salt,
		},
	}, nil
}

// VerifyPasswordHash verifies a salted SHA-512 PBKDF2 key derived from plaintext against ph.
// A return value of true indicates they match, false otherwise.
func VerifyPasswordHash(plaintext string, ph *PasswordHash) (bool, error) {
	if ph == nil {
		return false, errors.New("password hash is nil")
	}
	pb, err := pbkdf2.Key(sha512.New, plaintext, ph.SALTEDSHA512PBKDF2.Salt, int(ph.SALTEDSHA512PBKDF2.Iterations), sha512.Size)
	if err != nil {
		return false, fmt.Errorf("deriving PBKDF2 key for password verify: %w", err)
	}
	return 1 == subtle.ConstantTimeCompare(pb, ph.SALTEDSHA512PBKDF2.Entropy), nil
}
