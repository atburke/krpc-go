package api

import "fmt"

func (e *Error) Error() string {
	return fmt.Sprintf("%v - %v: %v", e.Service, e.Name, e.Description)
}
