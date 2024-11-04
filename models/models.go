package models

type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

// Define a struct that matches the JSON structure
type ResponseNodeIdentifier struct {
    Data int `json:"data"`
}