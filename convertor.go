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
	Channels   int
}

func NewConvertor(in, out *StreamInfo, resampleQuality int) (*Convertor, error) {
	if out.SampleRate <= 0 || in.SampleRate <= 0 {
		return nil, model.ErrInvalidSampleRate
	}
	if out.Format.FrameSize() < 0 || in.Format.FrameSize() < 0 {
		return nil, model.ErrInvalidFormat
	}
	if in.Channels <= 0 || out.Channels <= 0 {
		return nil, model.ErrInvalidChannels
	}

	if in.Channels != out.Channels {
		if in.Channels != 1 && out.Channels != 1 {
			return nil, model.ErrChannelsConvert
		}
	}

	formatConvertor, err := format.NewFormatConvertor(in.Format, out.Format, in.ByteOrder, out.ByteOrder)
	if err != nil {
		return nil, err
	}

	channels := out.Channels
	if in.Channels < out.Channels {
		channels = in.Channels
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
	var err error
	if p.out.Channels < p.in.Channels {
		data, err = StereoToMono(data, p.in.Format, p.in.Channels, p.in.ByteOrder)
		if err != nil {
			return nil, err
		}
	}

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
				return MonoToStereo(data2, p.out.Format, p.out.Channels)
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
				return MonoToStereo(data2, p.out.Format, p.out.Channels)
			}
		}
		return MonoToStereo(data1, p.out.Format, p.out.Channels)
	}
	if p.out.SampleRate != p.in.SampleRate {
		if p.out.ByteOrder == binary.LittleEndian {
			data2, err := p.resampler.Process(data)
			if err != nil {
				return nil, err
			}
			return MonoToStereo(data2, p.out.Format, p.out.Channels)
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
			return MonoToStereo(data2, p.out.Format, p.out.Channels)
		}
	}
	return MonoToStereo(data, p.out.Format, p.out.Channels)
}
