package passwd

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestPasswd(t *testing.T) {
	for i, p := range []string{"", "password", "hunter2", "Tr0ub4dor&3", "correct horse battery staple"} {
		t.Run(fmt.Sprintf("hash-%d-%d", i, len(p)), func(t *testing.T) {
			// make the hash
			ph, err := HashPassword(rand.Reader, p)
			if err != nil {
				t.Fatal(err)
			}

			// verify the password
			c, err := VerifyPasswordHash(p, ph)
			if err != nil {
				t.Fatal(err)
			}
			if c != true {
				t.Error("hashes do not match")
			}

			// now try an incorrect password
			c, err = VerifyPasswordHash("incorrect-password", ph)
			if err != nil {
				t.Fatal(err)
			}
			if c != false {
				t.Error("hashes should not match")
			}
		})
	}
}
