package pcm_convertor

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/resample"
)

func TestProcessor(t *testing.T) {
	f, err := os.Open("16k_16bit_mono.pcm")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	data16k16bit, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	outInfo := &StreamInfo{
		SampleRate: 32000,
		Format:     format.S32,
		ByteOrder:  binary.BigEndian,
		Channels:   1,
	}
	InInfo := &StreamInfo{
		SampleRate: 16000,
		Format:     format.S16,
		ByteOrder:  binary.LittleEndian,
		Channels:   1,
	}
	c, err := NewConvertor(InInfo, outInfo, resample.VeryHighQ)
	if err != nil {
		t.Fatal(err)
	}
	data32bit, err := c.Process(data16k16bit)
	if err != nil {
		t.Fatal(err)
	}
	nf, err := os.Create(fmt.Sprintf("%v_%v_%v.pcm", outInfo.Format.String(), outInfo.SampleRate, outInfo.ByteOrder))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	nf.Write(data32bit)
}
