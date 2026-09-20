// Command hash prints a bcrypt hash of the password on stdin, for NETRADOCK_PASSWORD_HASH.
// Use `make hash`, which prompts without echoing the password.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Cost 12 matches what the docs recommended for htpasswd; login runs once per session,
// so the extra work is unnoticeable while brute forcing stays expensive.
const cost = 12

func main() {
	in, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read password:", err)
		os.Exit(1)
	}
	password := strings.TrimRight(string(in), "\r\n")
	if password == "" {
		fmt.Fprintln(os.Stderr, "empty password")
		os.Exit(1)
	}
	out, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "hash password:", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
