package resample

/*
#cgo LDFLAGS: -lsoxr

#include <stdlib.h>
#include "soxr.h"
*/
import "C"
import (
	"errors"
	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/model"
	"runtime"
	"unsafe"
)

const (
	// Quality settings
	Quick     = 0 // Quick cubic interpolation
	LowQ      = 1 // LowQ 16-bit with larger rolloff
	MediumQ   = 2 // MediumQ 16-bit with medium rolloff
	HighQ     = 4 // High quality
	VeryHighQ = 6 // Very high quality
)

type Resampler struct {
	soxr     C.soxr_t
	inRate   int
	outRate  int
	channels int
	format   format.PcmFormat
}

var threads int

func init() {
	threads = runtime.NumCPU()
}

func NewResampler(inRate, outRate, channels, quality int, format format.PcmFormat) (*Resampler, error) {
	if inRate <= 0 || outRate <= 0 {
		return nil, model.ErrInvalidSampleRate
	}
	if format.ToSoxrDatatype() < 0 {
		return nil, model.ErrInvalidFormat
	}

	var soxr C.soxr_t
	var soxErr C.soxr_error_t
	ioSpec := C.soxr_io_spec(
		C.soxr_datatype_t(format.ToSoxrDatatype()),
		C.soxr_datatype_t(format.ToSoxrDatatype()),
	)
	qSpec := C.soxr_quality_spec(C.ulong(quality), 0)
	runtimeSpec := C.soxr_runtime_spec(C.uint(threads))
	soxr = C.soxr_create(
		C.double(float64(inRate)), C.double(float64(outRate)),
		C.uint(channels),
		&soxErr, &ioSpec, &qSpec, &runtimeSpec,
	)
	if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
		err := errors.New(C.GoString(soxErr))
		C.free(unsafe.Pointer(soxErr))
		return nil, err
	}
	C.free(unsafe.Pointer(soxErr))
	return &Resampler{
		soxr:     soxr,
		inRate:   inRate,
		outRate:  outRate,
		channels: channels,
		format:   format,
	}, nil
}

func (r *Resampler) Reset() (err error) {
	if r.soxr == nil {
		return errors.New("soxr resampler is nil")
	}
	C.soxr_clear(r.soxr)
	return
}

func (r *Resampler) Close() (err error) {
	if r.soxr == nil {
		return errors.New("soxr resampler is nil")
	}
	C.soxr_delete(r.soxr)
	r.soxr = nil
	return
}

func (r *Resampler) Process(data []byte) ([]byte, error) {
	if r.soxr == nil {
		return nil, errors.New("soxr resampler is nil")
	}
	if len(data) == 0 {
		return data, nil
	}
	if fragment := len(data) % (r.format.FrameSize() * r.channels); fragment != 0 {
		data = data[:len(data)-fragment]
	}
	framesLen := len(data) / r.format.FrameSize() / r.channels
	if framesLen == 0 {
		return nil, model.ErrFrameSizeError
	}
	framesOutLen := int(float64(framesLen) * (float64(r.outRate) / float64(r.inRate)))
	if framesOutLen == 0 {
		return []byte{}, nil
	}
	dataIn := C.CBytes(data)
	dataOut := C.malloc(C.size_t(framesOutLen * r.channels * r.format.FrameSize()))
	var soxErr C.soxr_error_t
	var read, done C.size_t = 0, 0
	defer func() {
		C.free(dataIn)
		C.free(dataOut)
		C.free(unsafe.Pointer(soxErr))
	}()

	for int(done) < framesOutLen {
		soxErr = C.soxr_process(r.soxr, C.soxr_in_t(dataIn), C.size_t(framesLen), &read, C.soxr_out_t(dataOut), C.size_t(framesOutLen), &done)
		if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
			err := errors.New(C.GoString(soxErr))
			if err != nil {
				return nil, err
			}
		}
		if int(read) == framesLen && int(done) < framesOutLen {
			// Indicate end of input to the resampler
			var d C.size_t = 0
			soxErr = C.soxr_process(r.soxr, C.soxr_in_t(nil), C.size_t(0), nil, C.soxr_out_t(dataOut), C.size_t(framesOutLen), &d)
			if C.GoString(soxErr) != "" && C.GoString(soxErr) != "0" {
				err := errors.New(C.GoString(soxErr))
				if err != nil {
					return nil, err
				}
			}
			done += d
			break
		}
	}
	return C.GoBytes(dataOut, C.int(int(done)*r.channels*r.format.FrameSize())), nil
}
