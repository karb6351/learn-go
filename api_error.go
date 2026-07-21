package main

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string { return e.Message }
