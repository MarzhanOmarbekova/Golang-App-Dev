package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type User struct {
	Name string `json:"name"`
	ID uint `json:"id"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
}

var Users = make([]User, 0)

func CreateUser(w http.ResponseWriter, r *http.Request)  {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := User{
		Name: req.Name,
		ID: vint(len(Users) + 1),
	}
	Users = append(Users, user)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	return
}

func GetUserbyID(w http.ResponseWriter, r *http.Response) {
	parts : = strings.Split(r.URL.path, "/")
	if len(parts)!= 3 {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	for _, user := range Users {
		if user.ID == uint(id) {
			err := json.NewEncoder(w).Encode(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
}

func GetUserByID ( w http.ResponseWriter, r *http.Response) {
	parts := strings.Split(r.URL.path, "/")
	if len(parts) != 3 {
		http.Error( w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	for _, user := range Users {
		if user.ID == uint(id){
			err := json.NewEncoder(w).Encode(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}
	http.Error(w, fmt.Sprintf("User %d not found", id), http.StatusNotFound)
}