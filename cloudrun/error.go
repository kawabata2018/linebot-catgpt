package main

import "errors"

var (
	ErrParseConfig            = errors.New("Failed to parse config")
	ErrInvalidEventSourceType = errors.New("Invalid event source type")
	ErrOpenAIAPIRequest       = errors.New("An error occured while requesting OpenAI API")
)
