package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/passwall/passwall-server/internal/app"
	"github.com/passwall/passwall-server/internal/storage"
	"github.com/passwall/passwall-server/model"
	"github.com/spf13/viper"

	"github.com/gorilla/mux"
)

const (
	loginDeleteSuccess = "Login deleted successfully!"
)

// FindAllLogins finds all logins
func FindAllLogins(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var loginList []model.Login

		// Setup variables
		transmissionKey := r.Context().Value("transmissionKey").(string)

		// Get all logins from db
		schema := r.Context().Value("schema").(string)
		loginList, err = s.Logins().All(schema)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Decrypt server side encrypted fields
		for i := range loginList {
			uLogin, err := app.DecryptModel(&loginList[i])
			if err != nil {
				RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
			loginList[i] = *uLogin.(*model.Login)
		}

		RespondWithEncJSON(w, http.StatusOK, transmissionKey, loginList)
	}
}

// FindLoginsByID finds a login by id
func FindLoginsByID(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Setup variables
		transmissionKey := r.Context().Value("transmissionKey").(string)

		// Check if id is integer
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Find login by id from db
		schema := r.Context().Value("schema").(string)
		login, err := s.Logins().FindByID(uint(id), schema)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Decrypt server side encrypted fields
		uLogin, err := app.DecryptModel(login)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Create DTO
		loginDTO := model.ToLoginDTO(uLogin.(*model.Login))

		RespondWithEncJSON(w, http.StatusOK, transmissionKey, loginDTO)
	}
}

// CreateLogin creates a login
func CreateLogin(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Setup variables
		env := viper.GetString("server.env")
		transmissionKey := r.Context().Value("transmissionKey").(string)

		// Update request body according to env.
		// If env is dev, then do nothing
		// If env is prod, then decrypt payload with transmission key
		if err := ToBody(r, env, transmissionKey); err != nil {
			RespondWithError(w, http.StatusBadRequest, InvalidRequestPayload)
			return
		}
		defer r.Body.Close()

		// Unmarshal request body to loginDTO
		var loginDTO model.LoginDTO
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&loginDTO); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		// Add new login to db
		schema := r.Context().Value("schema").(string)
		createdLogin, err := app.CreateLogin(s, &loginDTO, schema)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Decrypt server side encrypted fields
		decLogin, err := app.DecryptModel(createdLogin)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Create DTO
		createdLoginDTO := model.ToLoginDTO(decLogin.(*model.Login))

		RespondWithEncJSON(w, http.StatusOK, transmissionKey, createdLoginDTO)
	}
}

// UpdateLogin updates a login
func UpdateLogin(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Setup variables
		env := viper.GetString("server.env")
		transmissionKey := r.Context().Value("transmissionKey").(string)

		if err := ToBody(r, env, transmissionKey); err != nil {
			RespondWithError(w, http.StatusBadRequest, InvalidRequestPayload)
			return
		}
		defer r.Body.Close()

		// Unmarshal request body to loginDTO
		var loginDTO model.LoginDTO
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&loginDTO); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
			return
		}
		defer r.Body.Close()

		// Find login defined by id
		schema := r.Context().Value("schema").(string)
		login, err := s.Logins().FindByID(uint(id), schema)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Update login
		updatedLogin, err := app.UpdateLogin(s, login, &loginDTO, schema)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Decrypt server side encrypted fields
		decLogin, err := app.DecryptModel(updatedLogin)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Create DTO
		updatedLoginDTO := model.ToLoginDTO(decLogin.(*model.Login))

		RespondWithEncJSON(w, http.StatusOK, transmissionKey, updatedLoginDTO)
	}
}

// BulkUpdateLogins updates logins in payload
func BulkUpdateLogins(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginList []model.LoginDTO
		// var loginDTO model.LoginDTO

		// Setup variables
		env := viper.GetString("server.env")
		transmissionKey := r.Context().Value("transmissionKey").(string)
		if err := ToBody(r, env, transmissionKey); err != nil {
			RespondWithError(w, http.StatusBadRequest, InvalidRequestPayload)
			return
		}
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&loginList); err != nil {
			RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
		defer r.Body.Close()

		for _, loginDTO := range loginList {
			// Find login defined by id
			schema := r.Context().Value("schema").(string)
			login, err := s.Logins().FindByID(loginDTO.ID, schema)
			if err != nil {
				RespondWithError(w, http.StatusNotFound, err.Error())
				return
			}

			// Update login
			_, err = app.UpdateLogin(s, login, &loginDTO, schema)
			if err != nil {
				RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}
		}

		response := model.Response{
			Code:    http.StatusOK,
			Status:  "Success",
			Message: "Bulk update completed successfully!",
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

// DeleteLogin deletes a login
func DeleteLogin(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Find login defined by id
		schema := r.Context().Value("schema").(string)
		login, err := s.Logins().FindByID(uint(id), schema)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Delete login defined by id
		err = s.Logins().Delete(login.ID, schema)
		if err != nil {
			RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}

		// Generate response
		response := model.Response{
			Code:    http.StatusOK,
			Status:  Success,
			Message: loginDeleteSuccess,
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

// TestLogin login endpoint for test purposes
func TestLogin(s storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		response := model.Response{
			Code:    http.StatusOK,
			Status:  Success,
			Message: "Test success!",
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}
