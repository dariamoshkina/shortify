package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/dariamoshkina/shortify/internal"
	"github.com/google/uuid"
)

func WithUserCookie(k []byte) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		cookieFn := func(w http.ResponseWriter, r *http.Request) {
			userID := uuid.New()
			cookie, err := r.Cookie(internal.CookieNameUserID)

			if err != nil || cookie.Value == "" {
				encryptedCookie, err := encrypt(userID.String(), k)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{Name: internal.CookieNameUserID, Value: encryptedCookie, Path: "/", MaxAge: int(time.Hour)})
			} else {
				decryptedCookie, err := decrypt(cookie.Value, k)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				userID, err = uuid.Parse(decryptedCookie)
				if err != nil {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), internal.CtxKeyUserID{}, userID)
			r = r.WithContext(ctx)

			h.ServeHTTP(w, r)
		}

		return http.HandlerFunc(cookieFn)
	}
}

func encrypt(source string, key []byte) (string, error) {
	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	nonce, err := internal.RandomBytes(nonceSize)
	if err != nil {
		return "", err
	}

	encrypted := aesgcm.Seal(nonce, nonce, []byte(source), nil)

	return hex.EncodeToString(encrypted), nil
}

func decrypt(source string, key []byte) (string, error) {
	encrypted, err := hex.DecodeString(source)
	if err != nil {
		return "", err
	}

	aesblock, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(encrypted) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, data := encrypted[:nonceSize], encrypted[nonceSize:]

	decrypted, err := aesgcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
