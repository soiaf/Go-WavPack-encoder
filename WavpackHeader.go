package wvencode

/*
** WavpackHeader.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type WavpackHeader struct {
	ckID          [4]int
	ckSize        int // was uint32_t in C
	version       int
	track_no      int  // was uchar in C
	index_no      int  // was uchar in C
	total_samples uint // was uint32_t in C
	block_index   int  // was uint32_t in C
	block_samples int  // was uint32_t in C
	flags         uint // was uint32_t in C
	crc           uint
}
