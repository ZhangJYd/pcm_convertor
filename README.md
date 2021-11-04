Resample part relies on SOXR.

To install make sure you have libsoxr installed, then run:
```
go get -u github.com/ZhangJYd/pcm_convertor
```
example:

```go
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/ZhangJYd/pcm_convertor"
	"github.com/ZhangJYd/pcm_convertor/format"
	"github.com/ZhangJYd/pcm_convertor/resample"
)

func main() {
	f, err := os.Open("16k_16bit_mono.pcm")
	if err != nil {
		log.Println(err)
		return
	}

	defer f.Close()

	outInfo := &pcm_convertor.StreamInfo{
		SampleRate: 32000,
		Format:     format.F32,
		ByteOrder:  binary.BigEndian,
		Channels:   3,
	}
	inInfo := &pcm_convertor.StreamInfo{
		SampleRate: 16000,
		Format:     format.S16,
		ByteOrder:  binary.LittleEndian,
		Channels:   1,
	}
	c, err := pcm_convertor.NewConvertor(inInfo, outInfo, resample.Quick)
	if err != nil {
		log.Println(err)
		return
	}
	defer c.Close()
	outF, err := os.Create(
		fmt.Sprintf("%v_%v_%v_%vchannels.pcm",
			outInfo.Format.String(), outInfo.SampleRate, outInfo.ByteOrder, outInfo.Channels),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer outF.Close()

	chuckSize := 128
	for {
		chuck := make([]byte, inInfo.Format.FrameSize()*chuckSize)
		n, err := f.Read(chuck)
		if err != nil || n < inInfo.Format.FrameSize()*chuckSize {
			log.Println(err)
			break
		}
		stream, err := c.Process(chuck)
		if err != nil {
			log.Println(err)
			break
		}
		_, err = outF.Write(stream)
		if err != nil {
			log.Println(err)
			break
		}
	}
}


```
