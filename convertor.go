package pcm_convertor

import (
	"encoding/binary"

	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/model"
	"github.com/ZhangJYd/pcm_convertor/resample"
)

type Convertor struct {
	out *StreamInfo
	in  *StreamInfo

	formatConvertor *format.Convertor
	resampler       *resample.Resampler
}

type StreamInfo struct {
	SampleRate int
	Format     format.PcmFormat
	ByteOrder  binary.ByteOrder
}

func NewConvertor(in, out *StreamInfo, resampleQuality, channels int) (*Convertor, error) {
	if out.SampleRate <= 0 || in.SampleRate <= 0 {
		return nil, model.ErrInvalidSampleRate
	}
	if out.Format.FrameSize() < 0 || in.Format.FrameSize() < 0 {
		return nil, model.ErrInvalidFormat
	}
	if channels <= 0 {
		return nil, model.ErrInvalidChannels
	}

	formatConvertor, err := format.NewFormatConvertor(in.Format, out.Format, in.ByteOrder, out.ByteOrder)
	if err != nil {
		return nil, err
	}

	resampler, err := resample.NewResampler(in.SampleRate, out.SampleRate, channels, resampleQuality, out.Format)
	if err != nil {
		return nil, err
	}

	return &Convertor{
		out:             out,
		in:              in,
		formatConvertor: formatConvertor,
		resampler:       resampler,
	}, nil
}

func (p *Convertor) Close() error {
	return p.resampler.Close()
}

func (p *Convertor) Process(data []byte) ([]byte, error) {
	if p.out.Format != p.in.Format || p.out.ByteOrder != p.in.ByteOrder {
		data1, err := p.formatConvertor.Convert(data)
		if err != nil {
			return nil, err
		}
		if p.out.SampleRate != p.in.SampleRate {
			if p.out.ByteOrder == binary.LittleEndian {
				data2, err := p.resampler.Process(data1)
				if err != nil {
					return nil, err
				}
				return data2, nil
			} else {
				dataL, err := format.BigEndianLittleEndianConvert(data1, p.out.Format, binary.BigEndian, binary.LittleEndian)
				if err != nil {
					return nil, err
				}
				dataL1, err := p.resampler.Process(dataL)
				if err != nil {
					return nil, err
				}
				data2, err := format.BigEndianLittleEndianConvert(dataL1, p.out.Format, binary.LittleEndian, binary.BigEndian)
				if err != nil {
					return nil, err
				}
				return data2, nil
			}
		}
		return data1, nil
	}
	if p.out.SampleRate != p.in.SampleRate {
		if p.out.ByteOrder == binary.LittleEndian {
			data2, err := p.resampler.Process(data)
			if err != nil {
				return nil, err
			}
			return data2, nil
		} else {
			dataL, err := format.BigEndianLittleEndianConvert(data, p.out.Format, binary.BigEndian, binary.LittleEndian)
			if err != nil {
				return nil, err
			}
			dataL1, err := p.resampler.Process(dataL)
			if err != nil {
				return nil, err
			}
			data2, err := format.BigEndianLittleEndianConvert(dataL1, p.out.Format, binary.LittleEndian, binary.BigEndian)
			if err != nil {
				return nil, err
			}
			return data2, nil
		}
	}
	return data, nil
}
