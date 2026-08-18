package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/rockpanel/rockpanel/internal/core"
	"github.com/rockpanel/rockpanel/pkg/types"
	"golang.org/x/crypto/bcrypt"
)

var (
	store       *sessions.CookieStore
	sessionName = "rockpanel_session"
)

type contextKey string

const userCtxKey contextKey = "user"

func Init(secret string) {
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func CreateSession(w http.ResponseWriter, r *http.Request, user *types.User) error {
	session, _ := store.Get(r, sessionName)
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["role"] = user.Role
	return session.Save(r, w)
}

func GetUserFromSession(r *http.Request) (*types.User, error) {
	session, err := store.Get(r, sessionName)
	if err != nil {
		return nil, err
	}
	uid, ok := session.Values["user_id"].(int64)
	if !ok {
		return nil, http.ErrNoCookie
	}
	return core.GetUserByID(uid)
}

func Logout(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, sessionName)
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := GetUserFromSession(r)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, user)
		next(w, r.WithContext(ctx))
	}
}

func RequireAPIAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token := authHeader[7:]
			prefix := ""
			if len(token) >= 8 {
				prefix = token[:8]
			}
			dbToken, err := core.GetAPITokenByPrefix(prefix)
			if err != nil || dbToken == nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			if dbToken.ExpiresAt > 0 && dbToken.ExpiresAt < time.Now().Unix() {
				http.Error(w, `{"error":"token expired"}`, http.StatusUnauthorized)
				return
			}
			if !CheckPassword(token, dbToken.TokenHash) {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			core.UpdateAPITokenLastUsed(dbToken.ID)
			user, _ := core.GetUserByID(dbToken.UserID)
			if user != nil {
				ctx := context.WithValue(r.Context(), userCtxKey, user)
				next(w, r.WithContext(ctx))
				return
			}
		}

		http.Error(w, `{"error":"invalid authorization"}`, http.StatusUnauthorized)
	}
}

func RequireSessionOrAPI(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			RequireAPIAuth(next)(w, r)
			return
		}
		RequireSession(next)(w, r)
	}
}

func RequireRole(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return RequireSessionOrAPI(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			allowed := false
			for _, role := range roles {
				if user.Role == role {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next(w, r)
		})
	}
}

func GetUserFromContext(ctx context.Context) *types.User {
	user, _ := ctx.Value(userCtxKey).(*types.User)
	return user
}

func GenerateToken() (rawToken, prefix string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	rawToken = base64.URLEncoding.EncodeToString(b)
	prefix = rawToken[:8]
	return rawToken, prefix, nil
}