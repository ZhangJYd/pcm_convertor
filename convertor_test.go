package pcm_convertor

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"pcm_convertor/format"
	"pcm_convertor/resample"
	"testing"
)

func TestProcessor(t *testing.T) {
	f, err := os.Open("16k_16bit.pcm")
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
	}
	InInfo := &StreamInfo{
		SampleRate: 16000,
		Format:     format.S16,
		ByteOrder:  binary.LittleEndian,
	}
	c, err := NewConvertor(InInfo, outInfo, resample.VeryHighQ, 1)
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
