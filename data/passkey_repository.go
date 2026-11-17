package data

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dhruv2223/reeling-it/logger"
	"github.com/dhruv2223/reeling-it/models"
	"github.com/go-webauthn/webauthn/webauthn"
)

// PasskeyRepository manages WebAuthn passkey data using a database.
type PasskeyRepository struct {
	db       *sql.DB                         // Database connection
	mu       sync.Mutex                      // Guards sessions; handlers run concurrently
	sessions map[string]webauthn.SessionData // In-memory session storage
	log      logger.Logger                   // Logger for debugging and errors
}

// NewPasskeyRepository initializes a new PasskeyRepository with a database connection.
func NewPasskeyRepository(db *sql.DB, log logger.Logger) *PasskeyRepository {
	return &PasskeyRepository{
		db:       db,
		sessions: make(map[string]webauthn.SessionData),
		log:      log,
	}
}

// GenSessionID generates a random session ID.
func (r *PasskeyRepository) GenSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetSession retrieves session data from the in-memory map.
func (r *PasskeyRepository) GetSession(token string) (webauthn.SessionData, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	val, ok := r.sessions[token]
	r.log.Info(fmt.Sprintf("GetSession: found=%v", ok))
	return val, ok
}

// SaveSession stores session data in the in-memory map, dropping expired
// sessions so abandoned registration/login attempts don't accumulate.
func (r *PasskeyRepository) SaveSession(token string, data webauthn.SessionData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	for t, s := range r.sessions {
		if !s.Expires.IsZero() && s.Expires.Before(now) {
			delete(r.sessions, t)
		}
	}
	r.log.Info("SaveSession")
	r.sessions[token] = data
}

// DeleteSession removes session data from the in-memory map.
func (r *PasskeyRepository) DeleteSession(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.log.Info("DeleteSession")
	delete(r.sessions, token)
}

func (r *PasskeyRepository) GetUserByEmail(email string) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", email))

	// Check if user exists by email
	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}

	// Fetch user credentials from passkeys table
	rows, err := r.db.Query("SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}

	// Construct and return PasskeyUser
	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)), // Convert int ID to byte slice
		Name:        email,
		DisplayName: email,
		Credentials: credentials,
	}
	return &user, nil
}

func (r *PasskeyRepository) GetUserByID(id int) (*models.PasskeyUser, error) {
	r.log.Info(fmt.Sprintf("Get User: %v", id))

	// Check if user exists by id
	var userID int
	err := r.db.QueryRow("SELECT id FROM users WHERE id = $1", id).Scan(&userID)
	if err == sql.ErrNoRows {
		r.log.Error("Failed to find new user", err)
		return nil, err
	} else if err != nil {
		r.log.Error("Failed to query user", err)
		return nil, err
	}

	// Fetch user credentials from passkeys table
	rows, err := r.db.Query("SELECT keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query passkeys", err)
		return nil, err
	}
	defer rows.Close()

	// Fetch the email of the user
	var email string
	err = r.db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Error("Failed to find user email", err)
			return nil, err
		}
		r.log.Error("Failed to query user email", err)
		return nil, err
	}

	var credentials []webauthn.Credential
	for rows.Next() {
		var keys string
		if err := rows.Scan(&keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			return nil, err
		}
		cred, err := deserializeCredential(keys)
		if err != nil {
			r.log.Error("Failed to deserialize credential", err)
			continue // Skip invalid credentials
		}
		credentials = append(credentials, cred)
	}

	// Construct and return PasskeyUser
	user := models.PasskeyUser{
		ID:          []byte(strconv.Itoa(userID)), // Convert int ID to byte slice
		Name:        email,
		DisplayName: email,
		Credentials: credentials,
	}
	return &user, nil
}

// SaveUser updates the user's credentials in the database.
func (r *PasskeyRepository) SaveUser(user models.PasskeyUser) {
	r.log.Info(fmt.Sprintf("SaveUser: %v", user.WebAuthnName()))

	// Convert user ID from byte slice to integer
	userID, err := strconv.Atoi(string(user.ID))
	if err != nil {
		r.log.Error("Invalid user ID", err)
		return
	}

	// Map each stored credential ID to its row, so a credential whose data
	// changed (e.g. its sign counter after a login) is updated in place
	// instead of being inserted again as a duplicate row.
	rows, err := r.db.Query("SELECT id, keys FROM passkeys WHERE user_id = $1", userID)
	if err != nil {
		r.log.Error("Failed to query existing passkeys", err)
		return
	}
	existing := make(map[string]int)
	for rows.Next() {
		var rowID int
		var keys string
		if err := rows.Scan(&rowID, &keys); err != nil {
			r.log.Error("Failed to scan passkey row", err)
			continue
		}
		if cred, err := deserializeCredential(keys); err == nil {
			existing[string(cred.ID)] = rowID
		}
	}
	rows.Close()

	for _, cred := range user.Credentials {
		keys, err := serializeCredential(cred)
		if err != nil {
			r.log.Error("Failed to serialize credential", err)
			continue
		}

		if rowID, ok := existing[string(cred.ID)]; ok {
			if _, err = r.db.Exec("UPDATE passkeys SET keys = $1 WHERE id = $2", keys, rowID); err != nil {
				r.log.Error("Failed to update passkey", err)
			}
		} else {
			if _, err = r.db.Exec("INSERT INTO passkeys (user_id, keys) VALUES ($1, $2)", userID, keys); err != nil {
				r.log.Error("Failed to insert passkey", err)
			}
		}
	}
}

// serializeCredential converts a WebAuthn credential to a JSON string.
func serializeCredential(cred webauthn.Credential) (string, error) {
	data, err := json.Marshal(cred)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credential: %w", err)
	}
	return string(data), nil
}

// deserializeCredential converts a JSON string back to a WebAuthn credential.
func deserializeCredential(data string) (webauthn.Credential, error) {
	var cred webauthn.Credential
	err := json.Unmarshal([]byte(data), &cred)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("failed to unmarshal credential: %w", err)
	}
	return cred, nil
}
