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
	Key int `json:"key" binding:"required"`
}

// Response struct to format the response JSON
type FindSuccessorErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// Response struct to format the response JSON
type FindSuccessorSuccessResponse struct {
	Successor int    `json:"successor"`
	Message   string `json:"message,omitempty"`
}

type StoreDataRequest struct {
	Data string `json:"data" binding:"required"`
}

// Response struct to format the response JSON
type StoreDataResponsee struct {
	Message string `json:"message"`
	Key     int    `json:"key,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Response struct to format the response JSON
type InternalStoreDataRequest struct {
	Message string `json:"message"`
	Key     int    `json:"key,omitempty"`
	Data    string `json:"data,omitempty"`
}

// Response struct to format the response JSON
type InternalStoreDataResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type RetrieveDataResponse struct {
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type InternalRetrieveDataResponse struct {
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type NotifyRequest struct {
	Key         int    `json:"key"`
	NodeAddress string `json:"node_address"`
}

type NotifyResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type StablizationSuccessorRequest struct {
	Message     string `json:"message"`
	Key         int    `json:"key"`
	NodeAddress string `json:"node_address"`
}

type StablizationSuccessorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type UpdateMetadataUponNewNodeJoinRequest struct {
	Key         int    `json:"key"`
	NodeAddress string `json:"node_address"`
}

type UpdateMetadataUponNewNodeJoinResponse struct {
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type LeaveRingMessage struct {
	DepartingNodeID   int            `json:"departing_node_id"`
	Keys              map[int]string `json:"keys"`
	SuccessorListNode int            `json:"successor_list_node"` // last node in departing node's successor list to be added to target node's successor list
	NewSuccessor      int            `json:"new_successor"`
	NewPredecessor    int            `json:"new_predecessor"`
}

type CycleCheckMessage struct {
	Initiator int
	Nodes     []int
}

func NewCycleCheckMessage() *CycleCheckMessage {
	return &CycleCheckMessage{
		Initiator: -1,
		Nodes:     make([]int, 0),
	}
}

type HTTPErrorMessage struct {
	msg string
	err string
}

func NewHTTPErrorMessage(msg string, err string) *HTTPErrorMessage {
	return &HTTPErrorMessage{
		msg: msg,
		err: err,
	}
}
