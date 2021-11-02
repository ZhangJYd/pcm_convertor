package format

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"

	"github.com/ZhangJYd/pcm_convertor/model"
)

type Convertor struct {
	inF          PcmFormat
	outF         PcmFormat
	inByteOrder  binary.ByteOrder
	outByteOrder binary.ByteOrder
}

func (c *Convertor) Convert(data []byte) ([]byte, error) {
	if len(data) < c.inF.FrameSize() {
		return nil, model.ErrPcmLenError
	}
	if fragment := len(data) % c.inF.FrameSize(); fragment != 0 {
		data = data[:len(data)-fragment]
	}
	buf := new(bytes.Buffer)
	for i := 0; ; {
		if i+c.inF.FrameSize() > len(data) {
			return buf.Bytes(), nil
		}
		if c.inByteOrder == c.outByteOrder {
			err := ConvertFormatForFrame(data[i:i+c.inF.FrameSize()], buf, c.inF, c.outF, c.outByteOrder)
			if err != nil {
				return nil, err
			}
		} else {
			inData, err := BigEndianLittleEndianConvert(data[i:i+c.inF.FrameSize()], c.inF, c.inByteOrder, c.outByteOrder)
			if err != nil {
				return nil, err
			}
			err = ConvertFormatForFrame(inData, buf, c.inF, c.outF, c.outByteOrder)
			if err != nil {
				return nil, err
			}
		}
		i += c.inF.FrameSize()
	}
}

func NewFormatConvertor(inF, outF PcmFormat, inByteOrder, outByteOrder binary.ByteOrder) (*Convertor, error) {
	if inF.FrameSize() <= 0 || outF.FrameSize() <= 0 {
		return nil, model.ErrInvalidFormat
	}
	return &Convertor{
		inF:          inF,
		outF:         outF,
		inByteOrder:  inByteOrder,
		outByteOrder: outByteOrder,
	}, nil
}

func Float32ToInt32(data float32) int32 {
	p := float64(data) * float64(math.MaxInt32)
	if p > 2147483647 {
		return 2147483647
	}
	if p < -2147483648 {
		return math.MinInt32
	}
	return int32(p)
}

func Int32ToFloat32(data int32) float32 {
	return float32(data) / float32(math.MaxInt32)
}

func BytesToInt16(data []byte, byteOrder binary.ByteOrder) (int16, error) {
	if byteOrder == binary.LittleEndian {
		return int16(binary.LittleEndian.Uint16(data)), nil
	}
	if byteOrder == binary.BigEndian {
		return int16(binary.BigEndian.Uint16(data)), nil
	}
	return 0, model.ErrInvalidByteOrder
}

func BytesToInt32(data []byte, byteOrder binary.ByteOrder) (int32, error) {
	if byteOrder == binary.LittleEndian {
		return int32(binary.LittleEndian.Uint32(data)), nil
	}
	if byteOrder == binary.BigEndian {
		return int32(binary.BigEndian.Uint32(data)), nil
	}
	return 0, model.ErrInvalidByteOrder
}

func BytesToFloat32(data []byte, byteOrder binary.ByteOrder) (float32, error) {
	if byteOrder == binary.LittleEndian {
		return math.Float32frombits(binary.LittleEndian.Uint32(data)), nil
	}
	if byteOrder == binary.BigEndian {
		return math.Float32frombits(binary.BigEndian.Uint32(data)), nil
	}
	return 0, model.ErrInvalidByteOrder
}

func BytesToFloat64(data []byte, byteOrder binary.ByteOrder) (float64, error) {
	if byteOrder == binary.LittleEndian {
		return math.Float64frombits(binary.LittleEndian.Uint64(data)), nil
	}
	if byteOrder == binary.BigEndian {
		return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
	}
	return 0, model.ErrInvalidByteOrder
}

func Float32ToBytes(data float32, byteOrder binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, byteOrder, data)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func Float64ToBytes(data float64, byteOrder binary.ByteOrder) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, byteOrder, data)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

// ConvertFormatForFrame Single frame conversion
func ConvertFormatForFrame(inData []byte, outW io.Writer, inF, outF PcmFormat, order binary.ByteOrder) error {
	if len(inData) != inF.FrameSize() {
		return model.ErrFrameSizeError
	}

	if outW == nil {
		return errors.New("io.Writer is nil")
	}

	if inF == outF {
		_, err := outW.Write(inData)
		return err
	}

	if inF == F64 {
		f64, err := BytesToFloat64(inData, order)
		if err != nil {
			return err
		}
		return ConvertFormatForFrame(Float32ToBytes(float32(f64), order), outW, F32, outF, order)
	}

	if outF == F64 {
		buf := new(bytes.Buffer)
		err := ConvertFormatForFrame(inData, buf, inF, F32, order)
		if err != nil {
			return err
		}
		f32, err := BytesToFloat32(buf.Bytes(), order)
		if err != nil {
			return err
		}
		_, err = outW.Write(Float64ToBytes(float64(f32), order))
		if err != nil {
			return err
		}
	}

	if inF != F32 && outF != F32 {
		outData := make([]byte, outF.FrameSize())
		if order == binary.LittleEndian {
			for i := range outData {
				if i < len(inData) {
					outData[len(outData)-i-1] = inData[len(inData)-i-1]
				} else {
					break
				}
			}
			_, err := outW.Write(outData)
			return err
		} else if order == binary.BigEndian {
			copy(outData, inData)
			_, err := outW.Write(outData)
			return err
		}
	} else if inF == F32 {
		buf := new(bytes.Buffer)
		f32, err := BytesToFloat32(inData, order)
		if err != nil {
			return err
		}
		i32 := Float32ToInt32(f32)
		err = binary.Write(buf, order, i32)
		if err != nil {
			return err
		}
		return ConvertFormatForFrame(buf.Bytes(), outW, S32, outF, order)
	} else {
		in32Data := make([]byte, 4)
		if order == binary.LittleEndian {
			for i := range in32Data {
				if i < len(inData) {
					in32Data[len(in32Data)-i-1] = inData[len(inData)-i-1]
				} else {
					break
				}
			}
		} else if order == binary.BigEndian {
			copy(in32Data, inData)
		}
		in32, err := BytesToInt32(in32Data, order)
		if err != nil {
			return err
		}
		F32Data := Float32ToBytes(Int32ToFloat32(in32), order)
		_, err = outW.Write(F32Data)
		return err
	}
	return nil
}

func BigEndianLittleEndianConvert(data []byte, inF PcmFormat, inByteOrder, outByteOrder binary.ByteOrder) ([]byte, error) {
	if inByteOrder == outByteOrder || len(data) == 0 {
		return data, nil
	}
	if len(data) < inF.FrameSize() {
		return nil, model.ErrFrameSizeError
	}
	var err error
	buf := new(bytes.Buffer)
	for i := 0; i < len(data); {
		var m interface{}
		switch inF {
		case U8:
			return data, nil
		case S16:
			m, err = BytesToInt16(data[i:i+inF.FrameSize()], inByteOrder)
		case S32:
			m, err = BytesToInt32(data[i:i+inF.FrameSize()], inByteOrder)
		case F32:
			m, err = BytesToFloat32(data[i:i+inF.FrameSize()], inByteOrder)
		case F64:
			m, err = BytesToFloat64(data[i:i+inF.FrameSize()], inByteOrder)
		default:
			return nil, model.ErrInvalidFormat
		}
		if err != nil {
			return nil, err
		}
		err = binary.Write(buf, outByteOrder, m)
		if err != nil {
			return nil, err
		}
		i += inF.FrameSize()
	}
	return buf.Bytes(), nil
}
