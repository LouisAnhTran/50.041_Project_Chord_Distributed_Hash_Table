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
type FindSuccessorRequest struct {
    Key  int `json:"key" binding:"required"`
}

// Response struct to format the response JSON
type FindSuccessorErrorResponse struct {
    Message string `json:"message"`
    Error   string  `json:"error,omitempty"`
}

// Response struct to format the response JSON
type FindSuccessorSuccessResponse struct {
    Successor int `json:"successor"`
    Message   string  `json:"message,omitempty"`
}

type StoreDataRequest struct {
    Data  string `json:"data" binding:"required"`
}

// Response struct to format the response JSON
type StoreDataResponsee struct {
    Message string `json:"message"`
    Key string `json:"key,omitempty"`
    Error   string  `json:"error,omitempty"`
}