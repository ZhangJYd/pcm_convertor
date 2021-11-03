package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

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
		Format:     format.F32,
		ByteOrder:  binary.BigEndian,
	}
	InInfo := &pcm_convertor.StreamInfo{
		SampleRate: 16000,
		Format:     format.S16,
		ByteOrder:  binary.LittleEndian,
	}
	c, err := pcm_convertor.NewConvertor(InInfo, outInfo, resample.Quick, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()
	outF, err := os.Create(fmt.Sprintf("%v_%v_%v.pcm", outInfo.Format.String(), outInfo.SampleRate, outInfo.ByteOrder))
	if err != nil {
		log.Println(err)
		return
	}
	defer outF.Close()

	chuckSize := 128
	for {
		chuck := make([]byte, InInfo.Format.FrameSize()*chuckSize)
		n, err := f.Read(chuck)
		if err != nil || n < InInfo.Format.FrameSize()*chuckSize {
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
	log.Println(time.Since(t1) / 200)
}