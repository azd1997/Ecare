package net

import "errors"

var (
	ErrUnavailableNode = errors.New("unavailable node address")
	ErrSendToSelf = errors.New("cannot send data to self")
	ErrUnKnownNode = errors.New("unknown node address")
	ErrUnknownInvType = errors.New("unknown inventory type")
	ErrUnknownGetDataType = errors.New("unknown getdata type")
	ErrNoValidTransaction = errors.New("no valid transaction in txPool")
	ErrBlockAlreadyExists = errors.New("this block already exists in local chain")
	ErrInvalidBlock = errors.New("invalid block for local chain")
)
