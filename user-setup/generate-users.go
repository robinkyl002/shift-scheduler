package main

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserFile struct {
	Users []User `json:"users"`
}

func updateUserPassword(filename, username, plainTextPassword string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	var userFile UserFile
	err = json.Unmarshal(data, &userFile)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %v", err)
	}

	found := false
	for i := range userFile.Users {
		if userFile.Users[i].Username == username {
			userFile.Users[i].Password = string(hashedPassword)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("user '%s' not found", username)
	}

	updated, err := json.MarshalIndent(userFile, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling JSON: %v", err)
	}

	return os.WriteFile(filename, updated, 0644)
}
