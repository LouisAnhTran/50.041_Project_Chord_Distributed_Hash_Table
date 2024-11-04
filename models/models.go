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

// Request struct to bind the incoming JSON
type Request struct {
    ID  int `json:"id" binding:"required"`
}


// Response struct to format the response JSON
type Response struct {
    Message string `json:"message"`
    Data    *Request `json:"data,omitempty"`
    Error   string  `json:"error,omitempty"`
}