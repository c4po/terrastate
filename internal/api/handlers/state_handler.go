package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/c4po/terrastate/internal/models"
	"github.com/c4po/terrastate/internal/storage"
	"github.com/gorilla/mux"
)

type StateHandler struct {
	storage storage.StateStorage
}

func NewStateHandler(storage storage.StateStorage) *StateHandler {
	return &StateHandler{storage: storage}
}

func (h *StateHandler) GetState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workspace := vars["workspace"]
	id := vars["id"]

	state, err := h.storage.GetState(r.Context(), workspace, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(state)
}

func (h *StateHandler) PutState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	state := &models.State{
		ID:        vars["id"],
		Workspace: vars["workspace"],
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	state.State = body

	if err := h.storage.PutState(r.Context(), state); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *StateHandler) DeleteState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.storage.DeleteState(r.Context(), vars["workspace"], vars["id"]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *StateHandler) ListStates(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	states, err := h.storage.ListStates(r.Context(), vars["workspace"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(states)
}

func (h *StateHandler) Lock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var lock models.StateLock
	if err := json.NewDecoder(r.Body).Decode(&lock); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lock.Path = vars["workspace"] + "/" + vars["id"]
	if err := h.storage.Lock(r.Context(), &lock); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *StateHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := h.storage.Unlock(r.Context(), vars["workspace"], vars["id"]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Implement other handler methods...
