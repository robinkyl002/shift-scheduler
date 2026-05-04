package main

import (
	"fmt"
)

func main() {
	users := make(map[string]User)
	users["admin"] = User{Username: "admin", Password: "adminpass"}
	users["user1"] = User{Username: "user1", Password: "userpass"}

	for username, user := range users {
		err := updateUserPassword("users.json", username, user.Password)
		if err != nil {
			fmt.Printf("Error updating password for user '%s': %v\n", username, err)
		}
	}

	fmt.Println("Update process completed.")
}
