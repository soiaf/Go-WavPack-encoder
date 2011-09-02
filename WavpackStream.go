package wvencode

/*
** WavpackStream.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type WavpackStream struct {
	wphdr        WavpackHeader
	wvbits       Bitstream
	wvcbits      Bitstream
	dc           DeltaData
	w            WordsData
	blockbuff    []byte
	blockend     int
	block2buff   []byte
	block2end    int
	bits         int
	lossy_block  int
	num_terms    int
	sample_index int // was uint32_t in C

	decorr_passes [16]DecorrPass
}
