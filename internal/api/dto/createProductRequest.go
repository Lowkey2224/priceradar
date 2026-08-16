package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type CreateProductRequest struct {
	Title       string   `json:"title"`
	Urls        []string `json:"urls"`
	TargetPrice int64    `json:"targetPrice"`
}

func (req CreateProductRequest) Valid() error {
	errMsg := []byte{}
	if len(req.Urls) == 0 {
		errMsg = fmt.Append(errMsg, "Urls must contain Elements,")
	}
	if req.TargetPrice < 1 {
		errMsg = fmt.Append(errMsg, "targetPrice must be bigger than 0,")
	}
	if len(req.Title) == 0 {
		errMsg = fmt.Append(errMsg, "title must not be empty")
	}

	if len(errMsg) != 0 {
		return errors.New(string(errMsg))
	}
	return nil
}

func (req *CreateProductRequest) New(r io.Reader) error {
	return json.NewDecoder(r).Decode(&req)
}
