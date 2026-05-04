package main

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserFile struct {
	Users []User `json:"users"`
}

type TemplateData struct {
	IsAuthenticated bool
	Username        string
	Role            string
}
