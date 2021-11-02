package format

type PcmFormat uint32

const (
	U8 PcmFormat = iota
	S16
	S24
	S32
	F32
	F64
)

func (f *PcmFormat) FrameSize() int {
	switch *f {
	case U8:
		return 1
	case S16:
		return 2
	case S24:
		return 3
	case S32:
		return 4
	case F32:
		return 4
	case F64:
		return 8
	}
	return -1
}

func (f *PcmFormat) String() string {
	switch *f {
	case U8:
		return "unsigned-8-bit"
	case S16:
		return "signed-16-bit"
	case S24:
		return "signed-24-bit"
	case S32:
		return "signed-32-bit"
	case F32:
		return "32-bit-float"
	case F64:
		return "64-bit-float"
	}
	return "unknown format"
}

func (f *PcmFormat) ToSoxrDatatype() int {
	switch *f {
	case U8:
		return -1
	case S16:
		return 3
	case S24:
		return -1
	case S32:
		return 2
	case F32:
		return 0
	case F64:
		return 1
	}
	return -1
}
