/*
Copyright © 2022 Thomas von Dein

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"crypto/cipher"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	mathrand "math/rand"
	"os"
	"time"
	"unsafe"

	"golang.org/x/crypto/argon2"
	chapo "golang.org/x/crypto/chacha20poly1305"
)

const (
	SaltSize   = 32         // in bytes
	NonceSize  = 24         // in bytes. taken from aead.NonceSize()
	KeySize    = uint32(32) // KeySize is 32 bytes (256 bits).
	KeyTime    = uint32(5)
	KeyMemory  = uint32(1024 * 64) // KeyMemory in KiB. here, 64 MiB.
	KeyThreads = uint8(4)
	chunkSize  = 1024 * 32 // chunkSize in bytes. here, 32 KiB.

	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"

	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// via https://gist.github.com/dopey/c69559607800d2f2f90b1b1ed4e550fb
func AssertAvailablePRNG() {
	// Assert that a cryptographically secure PRNG is available.
	// Panic otherwise.
	buf := make([]byte, 1)

	_, err := io.ReadFull(cryptorand.Reader, buf)
	if err != nil {
		panic(fmt.Sprintf("crypto/rand is unavailable: Read() failed with %#v", err))
	}
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateSecureRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := cryptorand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateSecureRandomString(n int) (string, error) {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

// via:
// https://stackoverflow.com/a/31832326
func GenerateMathRandomString(n int) string {
	b := make([]byte, n)
	var src = mathrand.NewSource(time.Now().UnixNano())
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letters) {
			b[i] = letters[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func GetRandomKey() ([]byte, error) {
	password, err := GenerateSecureRandomBytes(int(chapo.KeySize))
	if err != nil {
		return nil, err
	}

	salt, err := GenerateSecureRandomBytes(chapo.NonceSizeX)
	if err != nil {
		return nil, err
	}

	key := argon2.IDKey(password, salt, KeyTime, KeyMemory, KeyThreads, chapo.KeySize)

	return key, nil
}

func Encrypt(c *Conf, filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	size := info.Size()

	outfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer outfile.Close()

	key, err := GetRandomKey()
	if err != nil {
		return err
	}

	aead, err := chapo.NewX(key)
	if err != nil {
		return err
	}

	for i := 0; i < c.count; i++ {
		for {
			if size < chunkSize {
				if err := EncryptChunk(aead, outfile, size); err != nil {
					return err
				}

				break
			}

			if err := EncryptChunk(aead, outfile, chunkSize); err != nil {
				return err
			}

			size = size - chunkSize

			if size <= 0 {
				break
			}
		}
	}

	return nil
}

func EncryptChunk(aead cipher.AEAD, file *os.File, size int64) error {
	chunk := make([]byte, size)
	nonce, err := GenerateSecureRandomBytes(int(chapo.NonceSizeX))
	if err != nil {
		return err
	}

	cipher := aead.Seal(nil, nonce, chunk, nil)

	n, err := file.Write(cipher[:size])
	if err != nil {
		return err
	}

	if int64(n) != size {
		return errors.New("invalid number of bytes written")
	}

	return nil
}
