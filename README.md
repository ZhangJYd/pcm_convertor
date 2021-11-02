To install make sure you have libsoxr installed, then run:
```
go get -u github.com/ZhangJYd/pcm_convertor
```
example:

```
package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/ZhangJYd/pcm_convertor"
	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/resample"
)

func main() {
	f, err := os.Open("16k_16bit.pcm")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	outInfo := &pcm_convertor.StreamInfo{
		SampleRate: 32000,
		Format:     format.S32,
		ByteOrder:  binary.BigEndian,
	}
	InInfo := &pcm_convertor.StreamInfo{
		SampleRate: 16000,
		Format:     format.S16,
		ByteOrder:  binary.LittleEndian,
	}
	c, err := pcm_convertor.NewConvertor(outInfo, InInfo, resample.VeryHighQ, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	nf, err := os.Create(fmt.Sprintf("%v_%v_%v.pcm", outInfo.Format.String(), outInfo.SampleRate, outInfo.ByteOrder))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer nf.Close()

	chuckSize := 100
	for {
		chuck := make([]byte, InInfo.Format.FrameSize()*chuckSize)
		_, err := nf.Read(chuck)
		if err != nil {
			fmt.Println(err)
			return
		}
		stream, err := c.Process(chuck)
		if err != nil {
			fmt.Println(err)
			return
		}
		nf.Write(stream)
	}

}
```
