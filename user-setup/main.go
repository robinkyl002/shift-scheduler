package main

import (
	"fmt"
)

func main() {
	users := make(map[string]User)
	users["admin1"] = User{Username: "admin1", Password: "adminpass"}
	users["student1"] = User{Username: "student1", Password: "userpass"}

	for username, user := range users {
		err := updateUserPassword("users.json", username, user.Password)
		if err != nil {
			fmt.Printf("Error updating password for user '%s': %v\n", username, err)
		}
	}

	fmt.Println("Update process completed.")
}
