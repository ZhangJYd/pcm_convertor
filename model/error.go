package model

import "errors"

var (
	ErrInvalidFormat     = errors.New("invalid format")
	ErrInvalidByteOrder  = errors.New("invalid byte order")
	ErrInvalidSampleRate = errors.New("invalid sample rate")
	ErrInvalidChannels   = errors.New("invalid channels")
	ErrFrameSizeError    = errors.New("frame size model")
	ErrPcmLenError       = errors.New("pcm len model")
)
