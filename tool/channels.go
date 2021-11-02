package tool

import (
	"bytes"
	"encoding/binary"
	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/model"
)

func StereoToMono(data []byte, inFormat format.PcmFormat, channels int, order binary.ByteOrder) ([]byte, error) {
	if inFormat.FrameSize() < 0 {
		return nil, model.ErrInvalidFormat
	}
	if len(data) == 0 {
		return data, nil
	}
	if len(data) < inFormat.FrameSize() {
		return nil, model.ErrFrameSizeError
	}
	if channels == 1 {
		return data, nil
	}
	if channels < 1 {
		return nil, model.ErrInvalidChannels
	}
	if fragment := len(data) % (inFormat.FrameSize() * channels); fragment != 0 {
		data = data[:len(data)-fragment]
	}
	mono := new(bytes.Buffer)

	for i := 0; i < len(data); {
		chunk := data[i : i+(inFormat.FrameSize()*channels)]
		var sum float64
		switch inFormat {
		case format.U8:
			for j := 0; j < len(chunk); {
				sum += float64(chunk[j])
				j += inFormat.FrameSize()
			}
			sum = sum / float64(channels)
			err := binary.Write(mono, order, uint8(sum))
			if err != nil {
				return nil, err
			}
		case format.S16:
			for j := 0; j < len(chunk); {
				n, err := format.BytesToInt16(chunk[j:j+inFormat.FrameSize()], order)
				if err != nil {
					return nil, err
				}
				sum += float64(n)
				j += inFormat.FrameSize()
			}
			err := binary.Write(mono, order, int16(sum))
			if err != nil {
				return nil, err
			}
		case format.S32:
			for j := 0; j < len(chunk); {
				n, err := format.BytesToInt32(chunk[j:j+inFormat.FrameSize()], order)
				if err != nil {
					return nil, err
				}
				sum += float64(n)
				j += inFormat.FrameSize()
			}
			err := binary.Write(mono, order, int32(sum))
			if err != nil {
				return nil, err
			}
		case format.F32:
			for j := 0; j < len(chunk); {
				n, err := format.BytesToFloat32(chunk[j:j+inFormat.FrameSize()], order)
				if err != nil {
					return nil, err
				}
				sum += float64(n)
				j += inFormat.FrameSize()
			}
			err := binary.Write(mono, order, float32(sum))
			if err != nil {
				return nil, err
			}
		case format.F64:
			for j := 0; j < len(chunk); {
				n, err := format.BytesToFloat64(chunk[j:j+inFormat.FrameSize()], order)
				if err != nil {
					return nil, err
				}
				sum += float64(n)
				j += inFormat.FrameSize()
			}
			err := binary.Write(mono, order, sum)
			if err != nil {
				return nil, err
			}
		}
		i += inFormat.FrameSize() * channels
	}
	return mono.Bytes(), nil
}
