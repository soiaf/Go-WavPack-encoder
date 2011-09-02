package wvencode

/*
** BitsUtils.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

////////////////////////// Bitstream functions ////////////////////////////////
// Open the specified BitStream using the specified buffer pointers. It is
// assumed that enough buffer space has been allocated for all data that will
// be written, otherwise an error will be generated.
func bs_open_write(bs *Bitstream, buffer_start int, buffer_end int) {
	bs.error = 0
	bs.sr = 0
	bs.bc = 0
	bs.buf_index = buffer_start
	bs.start_index = bs.buf_index
	bs.end = buffer_end
	bs.active = 1 // indicates that the bitstream is being used
}

// This function is only called from the putbit() and putbits() when
// the buffer is full, which is now flagged as an error.
func bs_wrap(bs *Bitstream) {
	bs.buf_index = bs.start_index
	bs.error = 1
}

// This function calculates the approximate number of bytes remaining in the
// bitstream buffer and can be used as an early-warning of an impending overflow.
func bs_remain_write(bs Bitstream) int {

	if bs.error > 0 {
		return (-1)
	}

	return bs.end - bs.buf_index
}

// This function forces a flushing write of the standard BitStream, and
// returns the total number of bytes written into the buffer.
func bs_close_write(wps *WavpackStream) int {
	var bs Bitstream = wps.wvbits
	var bytes_written int = 0

	if bs.error != 0 {
		return -1
	}

	for (bs.bc != 0) || (((bs.buf_index - bs.start_index) & 1) != 0) {
		putbit_1(wps)
		bs = wps.wvbits // as putbit_1 makes changes
	}

	bytes_written = bs.buf_index - bs.start_index

	return bytes_written
}

// This function forces a flushing write of the correction BitStream, and
// returns the total number of bytes written into the buffer.
func bs_close_correction_write(wps *WavpackStream) int {
	var bs Bitstream = wps.wvcbits
	var bytes_written int = 0

	if bs.error != 0 {
		return -1
	}

	for (bs.bc != 0) || (((bs.buf_index - bs.start_index) & 1) != 0) {
		putbit_correction_1(wps)
		bs = wps.wvcbits // as putbit_correction_1 makes changes
	}

	bytes_written = bs.buf_index - bs.start_index

	return bytes_written
}
