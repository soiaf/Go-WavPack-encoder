package wvencode

/*
** WavpackContext.hx
**
** Copyright (c) 2013 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

import (
	"os"
)

type WavpackContext struct {
	config             WavpackConfig
	stream             WavpackStream
	error_message      string
	Infile             os.File
	Outfile            *os.File
	Correction_outfile *os.File
	total_samples      uint // was uint32_t in C
	lossy_blocks       int
	wvc_flag           int
	block_samples      uint
	acc_samples        uint
	filelen            uint
	file2len           uint
	stream_version     int
	Byte_idx           int // holds the current buffer position for the input WAV data
}
