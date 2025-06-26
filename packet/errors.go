package packet

import "errors"

var (
	ErrInvalidReader   = errors.New("invalid reader")
	ErrSeqOverflow     = errors.New("seq overflow")
	ErrRouteOverflow   = errors.New("route overflow")
	ErrMessageTooLarge = errors.New("message too large")
	ErrInvalidMessage  = errors.New("invalid message")
)
