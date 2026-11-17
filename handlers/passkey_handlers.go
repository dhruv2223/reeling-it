package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dhruv2223/reeling-it/data"
	"github.com/dhruv2223/reeling-it/logger"
	"github.com/dhruv2223/reeling-it/models"
	"github.com/dhruv2223/reeling-it/token"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// The WebAuthn session cookie is only needed by the /api/passkey/* endpoints.
const passkeyCookieName = "sid"
const passkeyCookiePath = "/api/passkey"

type WebAuthnHandler struct {
	storage  data.PasskeyStore
	logger   *logger.Logger
	webauthn *webauthn.WebAuthn
}

func NewWebAuthnHandler(storage data.PasskeyStore, logger *logger.Logger, webauthn *webauthn.WebAuthn) *WebAuthnHandler {
	return &WebAuthnHandler{
		storage:  storage,
		logger:   logger,
		webauthn: webauthn,
	}
}

func (h *WebAuthnHandler) writeJSONResponse(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return err
	}
	return nil
}

func setSessionCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     passkeyCookieName,
		Value:    value,
		Path:     passkeyCookiePath,
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie deletes the cookie; Path must match the one it was set with.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     passkeyCookieName,
		Value:    "",
		Path:     passkeyCookiePath,
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// loadSession returns the WebAuthn session referenced by the request's cookie.
func (h *WebAuthnHandler) loadSession(r *http.Request) (string, webauthn.SessionData, bool) {
	sid, err := r.Cookie(passkeyCookieName)
	if err != nil {
		h.logger.Error("Couldn't get the cookie for the session", err)
		return "", webauthn.SessionData{}, false
	}
	session, ok := h.storage.GetSession(sid.Value)
	if !ok {
		h.logger.Error("Session not found or expired", nil)
		return "", webauthn.SessionData{}, false
	}
	return sid.Value, session, true
}

func (h *WebAuthnHandler) WebAuthnRegistrationBeginHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value("email").(string)
	if !ok {
		h.logger.Error("Unable to retrieve email", nil)
		http.Error(w, "Unable to retrieve email", http.StatusInternalServerError)
		return
	}
	user, err := h.storage.GetUserByEmail(email)
	if err != nil {
		h.logger.Error("Failed to find user", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Stop the browser from registering a second passkey on an authenticator
	// that already holds one for this account.
	exclusions := make([]protocol.CredentialDescriptor, 0, len(user.Credentials))
	for _, cred := range user.Credentials {
		exclusions = append(exclusions, cred.Descriptor())
	}

	options, session, err := h.webauthn.BeginRegistration(user, webauthn.WithExclusions(exclusions))
	if err != nil {
		h.logger.Error("Can't begin WebAuthn registration", err)
		http.Error(w, "Can't begin WebAuthn Registration", http.StatusInternalServerError)
		return
	}

	// Make a session key and store the sessionData values
	t, err := h.storage.GenSessionID()
	if err != nil {
		h.logger.Error("Can't generate session id", err)
		http.Error(w, "Can't begin WebAuthn Registration", http.StatusInternalServerError)
		return
	}
	h.storage.SaveSession(t, *session)
	setSessionCookie(w, t)

	h.writeJSONResponse(w, options)
}

func (h *WebAuthnHandler) WebAuthnRegistrationEndHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := r.Context().Value("email").(string)
	if !ok {
		h.logger.Error("Unable to retrieve email", nil)
		http.Error(w, "Unable to retrieve email", http.StatusInternalServerError)
		return
	}

	sid, session, ok := h.loadSession(r)
	if !ok {
		http.Error(w, "Registration session expired, please try again", http.StatusBadRequest)
		return
	}
	// A session is single-use, whatever the outcome.
	h.storage.DeleteSession(sid)
	clearSessionCookie(w)

	user, err := h.storage.GetUserByEmail(email)
	if err != nil {
		h.logger.Error("Failed to find user", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	credential, err := h.webauthn.FinishRegistration(user, session, r)
	if err != nil {
		h.logger.Error("Couldn't finish the WebAuthn registration", err)
		http.Error(w, "Couldn't finish registration", http.StatusBadRequest)
		return
	}

	// Store the credential object
	user.AddCredential(credential)
	h.storage.SaveUser(*user)

	h.writeJSONResponse(w, map[string]bool{"success": true})
}

func (h *WebAuthnHandler) WebAuthnAuthenticationBeginHandler(w http.ResponseWriter, r *http.Request) {
	type AuthenticationRequest struct {
		Email string `json:"email"`
	}
	var req AuthenticationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode authentication request", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := req.Email
	h.logger.Info("Finding user " + email)

	user, err := h.storage.GetUserByEmail(email) // Find the user
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "No account found for that email", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("Failed to find user", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	options, session, err := h.webauthn.BeginLogin(user)
	if err != nil {
		h.logger.Error("Couldn't start a login", err)
		http.Error(w, "No passkey is registered for this account", http.StatusNotFound)
		return
	}

	// Make a session key and store the sessionData values
	t, err := h.storage.GenSessionID()
	if err != nil {
		h.logger.Error("Couldn't create a session ID", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.storage.SaveSession(t, *session)
	setSessionCookie(w, t)

	h.writeJSONResponse(w, options)
}

func (h *WebAuthnHandler) WebAuthnAuthenticationEndHandler(w http.ResponseWriter, r *http.Request) {
	sid, session, ok := h.loadSession(r)
	if !ok {
		http.Error(w, "Login session expired, please try again", http.StatusBadRequest)
		return
	}
	// A session is single-use, whatever the outcome.
	h.storage.DeleteSession(sid)
	clearSessionCookie(w)

	userID, err := strconv.Atoi(string(session.UserID)) // Convert []byte to int
	if err != nil {
		h.logger.Error("Failed to convert UserID to int", err)
		http.Error(w, "Invalid session data", http.StatusBadRequest)
		return
	}
	user, err := h.storage.GetUserByID(userID) // Get the user
	if err != nil {
		h.logger.Error("Failed to find user", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	credential, err := h.webauthn.FinishLogin(user, session, r)
	if err != nil {
		h.logger.Error("Couldn't finish login", err)
		http.Error(w, "Passkey verification failed", http.StatusUnauthorized)
		return
	}

	// A sign counter that went backwards suggests a cloned authenticator.
	if credential.Authenticator.CloneWarning {
		h.logger.Error("Passkey clone warning for user "+user.Name, nil)
		http.Error(w, "Passkey verification failed", http.StatusUnauthorized)
		return
	}

	// If login was successful, persist the updated sign counter
	user.UpdateCredential(credential)
	h.storage.SaveUser(*user)

	type PasskeyResponse struct {
		Success bool        `json:"success"`
		JWT     token.Token `json:"jwt"`
	}
	h.logger.Info("Sending JWT for " + user.Name)
	response := PasskeyResponse{
		Success: true,
		JWT:     token.CreateJWT(models.User{Email: user.Name}, *h.logger),
	}

	h.writeJSONResponse(w, response)
}
