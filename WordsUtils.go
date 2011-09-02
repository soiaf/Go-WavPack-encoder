package wvencode

/*
** WordsUtils.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */


//////////////////////////////// local macros /////////////////////////////////
const LIMIT_ONES = 16 // maximum consecutive 1s sent for "div" data

// these control the time constant "slow_level" which is used for hybrid mode
// that controls bitrate as a function of residual level (HYBRID_BITRATE).
const SLS = 8
const SLO = (1 << (SLS - 1))

// these control the time constant of the 3 median level breakpoints
const DIV0 = 128 // 5/7 of samples
const DIV1 = 64  // 10/49 of samples
const DIV2 = 32  // 20/343 of samples


///////////////////////////// local table storage ////////////////////////////
var bitset = [...]uint{
	1 << 0, 1 << 1, 1 << 2, 1 << 3, 1 << 4, 1 << 5, 1 << 6, 1 << 7, 1 << 8, 1 << 9,
	1 << 10, 1 << 11, 1 << 12, 1 << 13, 1 << 14, 1 << 15, 1 << 16, 1 << 17, 1 << 18,
	1 << 19, 1 << 20, 1 << 21, 1 << 22, 1 << 23, 1 << 24, 1 << 25, 1 << 26, 1 << 27,
	1 << 28, 1 << 29, 1 << 30, 1 << 31}

var bitmask = [...]uint{
	(1 << 0) - 1, (1 << 1) - 1, (1 << 2) - 1, (1 << 3) - 1, (1 << 4) - 1, (1 << 5) -
		1, (1 << 6) - 1, (1 << 7) - 1, (1 << 8) - 1, (1 << 9) - 1, (1 << 10) - 1,
	(1 << 11) - 1, (1 << 12) - 1, (1 << 13) - 1, (1 << 14) - 1, (1 << 15) - 1,
	(1 << 16) - 1, (1 << 17) - 1, (1 << 18) - 1, (1 << 19) - 1, (1 << 20) - 1,
	(1 << 21) - 1, (1 << 22) - 1, (1 << 23) - 1, (1 << 24) - 1, (1 << 25) - 1,
	(1 << 26) - 1, (1 << 27) - 1, (1 << 28) - 1, (1 << 29) - 1, (1 << 30) - 1,
	0x7fffffff}

var nbits_table = [...]int{
	0, 1, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, // 0 - 15
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, // 16 - 31
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, // 32 - 47
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, // 48 - 63
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 64 - 79
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 80 - 95
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 96 - 111
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, // 112 - 127
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 128 - 143
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 144 - 159
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 160 - 175
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 176 - 191
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 192 - 207
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 208 - 223
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, // 224 - 239
	8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8, 8}

var log2_table = [...]int{
	0x00, 0x01, 0x03, 0x04, 0x06, 0x07, 0x09, 0x0a, 0x0b, 0x0d, 0x0e, 0x10, 0x11, 0x12, 0x14,
	0x15, 0x16, 0x18, 0x19, 0x1a, 0x1c, 0x1d, 0x1e, 0x20, 0x21, 0x22, 0x24, 0x25, 0x26, 0x28,
	0x29, 0x2a, 0x2c, 0x2d, 0x2e, 0x2f, 0x31, 0x32, 0x33, 0x34, 0x36, 0x37, 0x38, 0x39, 0x3b,
	0x3c, 0x3d, 0x3e, 0x3f, 0x41, 0x42, 0x43, 0x44, 0x45, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4d,
	0x4e, 0x4f, 0x50, 0x51, 0x52, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5c, 0x5d, 0x5e,
	0x5f, 0x60, 0x61, 0x62, 0x63, 0x64, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e,
	0x6f, 0x70, 0x71, 0x72, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e,
	0x7f, 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d,
	0x8e, 0x8f, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9b,
	0x9c, 0x9d, 0x9e, 0x9f, 0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xa9,
	0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0, 0xb1, 0xb2, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7,
	0xb8, 0xb9, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf, 0xc0, 0xc0, 0xc1, 0xc2, 0xc3, 0xc4,
	0xc5, 0xc6, 0xc6, 0xc7, 0xc8, 0xc9, 0xca, 0xcb, 0xcb, 0xcc, 0xcd, 0xce, 0xcf, 0xd0, 0xd0,
	0xd1, 0xd2, 0xd3, 0xd4, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdc,
	0xdd, 0xde, 0xdf, 0xe0, 0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe4, 0xe5, 0xe6, 0xe7, 0xe7, 0xe8,
	0xe9, 0xea, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xee, 0xef, 0xf0, 0xf1, 0xf1, 0xf2, 0xf3, 0xf4,
	0xf4, 0xf5, 0xf6, 0xf7, 0xf7, 0xf8, 0xf9, 0xf9, 0xfa, 0xfb, 0xfc, 0xfc, 0xfd, 0xfe, 0xff,
	0xff}

var exp2_table = [...]int{
	0x00, 0x01, 0x01, 0x02, 0x03, 0x03, 0x04, 0x05, 0x06, 0x06, 0x07, 0x08, 0x08, 0x09, 0x0a,
	0x0b, 0x0b, 0x0c, 0x0d, 0x0e, 0x0e, 0x0f, 0x10, 0x10, 0x11, 0x12, 0x13, 0x13, 0x14, 0x15,
	0x16, 0x16, 0x17, 0x18, 0x19, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1d, 0x1e, 0x1f, 0x20, 0x20,
	0x21, 0x22, 0x23, 0x24, 0x24, 0x25, 0x26, 0x27, 0x28, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2c,
	0x2d, 0x2e, 0x2f, 0x30, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x35, 0x36, 0x37, 0x38, 0x39,
	0x3a, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f, 0x40, 0x41, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46,
	0x47, 0x48, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f, 0x50, 0x51, 0x51, 0x52, 0x53,
	0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5e, 0x5f, 0x60, 0x61,
	0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f, 0x70,
	0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
	0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
	0x90, 0x91, 0x92, 0x93, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9f, 0xa0,
	0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xaf, 0xb0, 0xb1,
	0xb2, 0xb3, 0xb4, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbc, 0xbd, 0xbe, 0xbf, 0xc0, 0xc2, 0xc3,
	0xc4, 0xc5, 0xc6, 0xc8, 0xc9, 0xca, 0xcb, 0xcd, 0xce, 0xcf, 0xd0, 0xd2, 0xd3, 0xd4, 0xd6,
	0xd7, 0xd8, 0xd9, 0xdb, 0xdc, 0xdd, 0xde, 0xe0, 0xe1, 0xe2, 0xe4, 0xe5, 0xe6, 0xe8, 0xe9,
	0xea, 0xec, 0xed, 0xee, 0xf0, 0xf1, 0xf2, 0xf4, 0xf5, 0xf6, 0xf8, 0xf9, 0xfa, 0xfc, 0xfd,
	0xff}

var ones_count_table = [...]int{
	0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1,
	0, 5, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2,
	0, 1, 0, 6, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1,
	0, 2, 0, 1, 0, 5, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1, 0, 3,
	0, 1, 0, 2, 0, 1, 0, 7, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2, 0, 1,
	0, 3, 0, 1, 0, 2, 0, 1, 0, 5, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1, 0, 2,
	0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 6, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4, 0, 1,
	0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 5, 0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 4,
	0, 1, 0, 2, 0, 1, 0, 3, 0, 1, 0, 2, 0, 1, 0, 8}

// this macro retrieves the specified median breakpoint (without frac; min = 1)
func GET_MED(wps *WavpackStream, med int, channel int) int {
	return (((wps.w.median[med][channel]) >> 4) + 1)
}

// These macros update the specified median breakpoints. Note that the median
// is incremented when the sample is higher than the median, else decremented.
// They are designed so that the median will never drop below 1 and the value
// is essentially stationary if there are 2 increments for every 5 decrements.
func INC_MED0(wps *WavpackStream, channel int) {
	wps.w.median[0][channel] += (((wps.w.median[0][channel] + DIV0) / DIV0) * 5)
}

func DEC_MED0(wps *WavpackStream, channel int) {
	wps.w.median[0][channel] -= (((wps.w.median[0][channel] + (DIV0 - 2)) / DIV0) * 2)
}

func INC_MED1(wps *WavpackStream, channel int) {
	wps.w.median[1][channel] += (((wps.w.median[1][channel] + DIV1) / DIV1) * 5)
}

func DEC_MED1(wps *WavpackStream, channel int) {
	wps.w.median[1][channel] -= (((wps.w.median[1][channel] + (DIV1 - 2)) / DIV1) * 2)
}

func INC_MED2(wps *WavpackStream, channel int) {
	wps.w.median[2][channel] += (((wps.w.median[2][channel] + DIV2) / DIV2) * 5)
}

func DEC_MED2(wps *WavpackStream, channel int) {
	wps.w.median[2][channel] -= (((wps.w.median[2][channel] + (DIV2 - 2)) / DIV2) * 2)
}

func count_bits(av uint) int {
	if av < (1 << 8) {
		return nbits_table[av]
	} else {
		if av < (1 << 16) {
			return nbits_table[(int)(av>>8)] + 8
		} else {
			if av < (1 << 24) {
				return nbits_table[(int)(av>>16)] + 16
			} else {
				return nbits_table[(int)(av>>24)] + 24
			}
		}
	}
	return 0 // not reached, but keeps compiler happy
}

func init_words(wps *WavpackStream) {
	if (wps.wphdr.flags & HYBRID_FLAG) > 0 {
		word_set_bitrate(wps)
	}
}

// Set up parameters for hybrid mode based on header flags and "bits" field.
// This is currently only set up for the HYBRID_BITRATE mode in which the
// allowed error varies with the residual level (from "slow_level"). The
// simpler mode (which is not used yet) has the error level directly
// controlled from the metadata.
func word_set_bitrate(wps *WavpackStream) {
	var bitrate_0 int = 0
	var bitrate_1 int = 0

	if (wps.wphdr.flags & HYBRID_BITRATE) > 0 {
		if wps.bits < 568 {
			bitrate_0 = 0
		} else {
			bitrate_0 = wps.bits - 568
		}

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			if (wps.wphdr.flags & HYBRID_BALANCE) > 0 {
				if (wps.wphdr.flags & JOINT_STEREO) != 0 {
					bitrate_1 = 256
				} else {
					bitrate_1 = 0
				}
			} else {
				bitrate_1 = bitrate_0

				if (wps.wphdr.flags & JOINT_STEREO) != 0 {
					if bitrate_0 < 128 {
						bitrate_1 += bitrate_0
						bitrate_0 = 0
					} else {
						bitrate_0 -= 128
						bitrate_1 += 128
					}
				}
			}
		}
	} else {
		bitrate_1 = 0
		bitrate_0 = 0
	}

	wps.w.bitrate_acc[0] = uint(bitrate_0 << 16)
	wps.w.bitrate_acc[1] = uint(bitrate_1 << 16)
}


// Allocates the correct space in the metadata structure and writes the
// current median values to it. Values are converted from 32-bit unsigned
// to our internal 16-bit mylog2 values, and read_entropy_vars () is called
// to read the values back because we must compensate for the loss through
// the log function.
func write_entropy_vars(wps *WavpackStream, wpmd *WavpackMetadata) {
	var byteptr []byte
	var byte_idx int = 0
	var temp int = 0

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data

	wpmd.id = ID_ENTROPY_VARS

	temp = mylog2(int(wps.w.median[0][0]))
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++

	temp = mylog2(int(wps.w.median[1][0]))
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++
	temp = mylog2(int(wps.w.median[2][0]))
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
		temp = mylog2(int(wps.w.median[0][1]))
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
		temp = mylog2(int(wps.w.median[1][1]))
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
		temp = mylog2(int(wps.w.median[2][1]))
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
	}

	wpmd.byte_length = byte_idx

	wpmd.data = byteptr

	read_entropy_vars(wps, wpmd)
}


// Allocates enough space in the metadata structure and writes the current
// high word of the bitrate accumulator and the slow_level values to it. The
// slow_level values are converted from 32-bit unsigned to our internal 16-bit
// mylog2 values. Afterward, read_entropy_vars () is called to read the values
// back because we must compensate for the loss through the log function and
// the truncation of the bitrate.
func write_hybrid_profile(wps *WavpackStream, wpmd *WavpackMetadata) {
	var byteptr []byte
	var byte_idx int = 0
	var temp int

	word_set_bitrate(wps)

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data

	wpmd.id = ID_HYBRID_PROFILE

	if (wps.wphdr.flags & HYBRID_BITRATE) != 0 {
		temp = log2s((int)(wps.w.slow_level[0]))
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			temp = log2s((int)(wps.w.slow_level[1]))
			byteptr[byte_idx] = byte(temp)
			byte_idx++
			byteptr[byte_idx] = byte(temp >> 8)
			byte_idx++
		}
	}

	temp = (int)(wps.w.bitrate_acc[0] >> 16)
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
		temp = (int)(wps.w.bitrate_acc[1] >> 16)
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
	}

	if (wps.w.bitrate_delta[0] | wps.w.bitrate_delta[1]) != 0 {
		temp = log2s((int)(wps.w.bitrate_delta[0]))
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			temp = log2s((int)(wps.w.bitrate_delta[1]))
			byteptr[byte_idx] = byte(temp)
			byte_idx++
			byteptr[byte_idx] = byte(temp >> 8)
			byte_idx++
		}
	}

	wpmd.byte_length = byte_idx
	read_hybrid_profile(wps, wpmd)

	wpmd.data = byteptr
}


// Read the median log2 values from the specifed metadata structure, convert
// them back to 32-bit unsigned values and store them. If length is not
// exactly correct then we flag and return an error
func read_entropy_vars(wps *WavpackStream, wpmd *WavpackMetadata) int {
	var byteptr []byte = wpmd.data
	var bytelengthcheck int

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) != 0 {
		bytelengthcheck = 6
	} else {
		bytelengthcheck = 12
	}

	if wpmd.byte_length != bytelengthcheck {
		return FALSE
	}

	wps.w.median[0][0] = int(exp2s((int(byteptr[0]) & 0xff) + ((int(byteptr[1]) & 0xff) << 8)))
	wps.w.median[1][0] = int(exp2s((int(byteptr[2]) & 0xff) + ((int(byteptr[3]) & 0xff) << 8)))
	wps.w.median[2][0] = int(exp2s((int(byteptr[4]) & 0xff) + ((int(byteptr[5]) & 0xff) << 8)))

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
		wps.w.median[0][1] = int(exp2s((int(byteptr[6]) & 0xff) + ((int(byteptr[7]) & 0xff) << 8)))
		wps.w.median[1][1] = int(exp2s((int(byteptr[8]) & 0xff) + ((int(byteptr[9]) & 0xff) << 8)))
		wps.w.median[2][1] = int(exp2s((int(byteptr[10]) & 0xff) + ((int(byteptr[11]) & 0xff) << 8)))
	}

	return TRUE
}

// Read the hybrid related values from the specifed metadata structure, convert
// them back to their internal formats and store them. The extended profile
// stuff is not implemented yet, so return an error if we get more data than
// we know what to do with.
func read_hybrid_profile(wps *WavpackStream, wpmd *WavpackMetadata) int {
	var byteptr []byte = wpmd.data
	var byte_idx int = 0

	if (wps.wphdr.flags & HYBRID_BITRATE) != 0 {
		wps.w.slow_level[0] = exp2s(int(byteptr[byte_idx]&0xff) +
			(int(byteptr[byte_idx+1]&0xff) << 8))
		byte_idx += 2

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			wps.w.slow_level[1] = exp2s(int(byteptr[byte_idx]&0xff) + (int(byteptr[byte_idx+1]&0xff) << 8))
			byte_idx += 2
		}
	}

	wps.w.bitrate_acc[0] = uint((int(byteptr[byte_idx]&0xff) + (int(byteptr[byte_idx+1]&0xff) << 8)) << 16)
	byte_idx += 2

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
		wps.w.bitrate_acc[1] = uint((int(byteptr[byte_idx]&0xff) + (int(byteptr[byte_idx+1]&0xff) << 8)) << 16)
		byte_idx += 2
	}

	if byte_idx < wpmd.byte_length {
		wps.w.bitrate_delta[0] = exp2s((int(byteptr[byte_idx]&0xff) + (int(byteptr[byte_idx+1]&0xff) << 8)))
		byte_idx += 2

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			wps.w.bitrate_delta[1] = exp2s((int(byteptr[byte_idx]&0xff) + (int(byteptr[byte_idx+1]&0xff) << 8)))
			byte_idx += 2
		}

		if byte_idx < wpmd.byte_length {
			return FALSE
		}
	} else {
		wps.w.bitrate_delta[1] = 0
		wps.w.bitrate_delta[0] = 0
	}

	return TRUE
}

// This function is called during both encoding and decoding of hybrid data to
// update the "error_limit" variable which determines the maximum sample error
// allowed in the main bitstream. In the HYBRID_BITRATE mode (which is the only
// currently implemented) this is calculated from the slow_level values and the
// bitrate accumulators. Note that the bitrate accumulators can be channelging.
func update_error_limit(wps *WavpackStream) {

	wps.w.bitrate_acc[0] += uint(wps.w.bitrate_delta[0])

	var bitrate_0 int = int(wps.w.bitrate_acc[0] >> 16)

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) != 0 {
		if (wps.wphdr.flags & HYBRID_BITRATE) != 0 {
			var slow_log_0 int = (int)((wps.w.slow_level[0] + SLO) >> SLS)

			if (slow_log_0 - bitrate_0) > -0x100 {
				wps.w.error_limit[0] = exp2s(slow_log_0 - bitrate_0 + 0x100)
			} else {
				wps.w.error_limit[0] = 0
			}
		} else {
			wps.w.error_limit[0] = exp2s(bitrate_0)
		}
	} else {
		var bitrate_1 int = 0

		wps.w.bitrate_acc[1] += uint(wps.w.bitrate_delta[1])
		bitrate_1 = (int)(wps.w.bitrate_acc[1] >> 16)

		if (wps.wphdr.flags & HYBRID_BITRATE) != 0 {
			var slow_log_0 int = (int)((wps.w.slow_level[0] + SLO) >> SLS)
			var slow_log_1 int = (int)((wps.w.slow_level[1] + SLO) >> SLS)

			if (wps.wphdr.flags & HYBRID_BALANCE) != 0 {
				var balance int = (slow_log_1 - slow_log_0 + bitrate_1 + 1) >> 1

				if balance > bitrate_0 {
					bitrate_1 = bitrate_0 * 2
					bitrate_0 = 0
				} else if -balance > bitrate_0 {
					bitrate_0 = bitrate_0 * 2
					bitrate_1 = 0
				} else {
					bitrate_1 = bitrate_0 + balance
					bitrate_0 = bitrate_0 - balance
				}
			}

			if (slow_log_0 - bitrate_0) > -0x100 {
				wps.w.error_limit[0] = exp2s(slow_log_0 - bitrate_0 + 0x100)
			} else {
				wps.w.error_limit[0] = 0
			}

			if (slow_log_1 - bitrate_1) > -0x100 {
				wps.w.error_limit[1] = exp2s(slow_log_1 - bitrate_1 + 0x100)
			} else {
				wps.w.error_limit[1] = 0
			}
		} else {
			wps.w.error_limit[0] = exp2s(bitrate_0)
			wps.w.error_limit[1] = exp2s(bitrate_1)
		}
	}
}


// This function writes the specified word to the open bitstream "wvbits" and,
// if the bitstream "wvcbits" is open, writes any correction data there. This
// function will work for either lossless or hybrid but because a version
// optimized for lossless exits below, it would normally be used for the hybrid
// mode only. The return value is the actual value stored to the stream (even
// if a correction file is being created) and is used as feedback to the
// predictor.
func send_word(wps *WavpackStream, value int, channel int) int {
	var sign int
	if value < 0 {
		sign = 1
	} else {
		sign = 0
	}

	var ones_count uint
	var low uint
	var high uint
	var mid uint

	if ((wps.w.median[0][0] & ^1) == 0) && (wps.w.holding_zero == 0) &&
		((wps.w.median[0][1] & ^1) == 0) {
		if wps.w.zeros_acc != 0 {
			if value != 0 {
				flush_word(wps)
			} else {
				wps.w.slow_level[channel] -= ((wps.w.slow_level[channel] + SLO) >> SLS)
				wps.w.zeros_acc++

				return 0
			}
		} else if value != 0 {
			putbit_0(wps)
		} else {
			wps.w.slow_level[channel] -= ((wps.w.slow_level[channel] + SLO) >> SLS)
			wps.w.median[0][0] = 0
			wps.w.median[1][0] = 0
			wps.w.median[2][0] = 0
			wps.w.median[0][1] = 0
			wps.w.median[1][1] = 0
			wps.w.median[2][1] = 0
			wps.w.zeros_acc = 1

			return 0
		}
	}

	if sign != 0 {
		value = ^value
	}

	if ((wps.wphdr.flags & HYBRID_FLAG) != 0) && (channel == 0) {
		update_error_limit(wps)
	}

	if value < GET_MED(wps, 0, channel) {
		low = 0
		ones_count = 0

		high = uint(GET_MED(wps, 0, channel) - 1)
		DEC_MED0(wps, channel)
	} else {
		low = uint(GET_MED(wps, 0, channel))
		INC_MED0(wps, channel)

		if (value - int(low)) < GET_MED(wps, 1, channel) {
			ones_count = 1

			high = (low + uint(GET_MED(wps, 1, channel))) - 1
			DEC_MED1(wps, channel)
		} else {
			low += uint(GET_MED(wps, 1, channel))
			INC_MED1(wps, channel)

			if (value - int(low)) < GET_MED(wps, 2, channel) {
				ones_count = 2

				high = (low + uint(GET_MED(wps, 2, channel))) - 1
				DEC_MED2(wps, channel)
			} else {
				ones_count = 2 + uint((value-int(low))/GET_MED(wps, 2, channel))
				low += ((ones_count - 2) * uint(GET_MED(wps, 2, channel)))

				high = (low + uint(GET_MED(wps, 2, channel))) - 1
				INC_MED2(wps, channel)
			}
		}
	}

	mid = (high + low + 1) >> 1

	if wps.w.holding_zero != 0 {
		if ones_count != 0 {
			wps.w.holding_one++
		}

		flush_word(wps)

		if ones_count != 0 {
			wps.w.holding_zero = 1
			ones_count--
		} else {
			wps.w.holding_zero = 0
		}
	} else {
		wps.w.holding_zero = 1
	}

	wps.w.holding_one = ones_count * 2

	if wps.w.error_limit[channel] == 0 {
		if high != low {
			var maxcode uint = high - low
			var code uint = uint(value) - low
			var bitcount int = count_bits(maxcode)
			var extras uint = bitset[bitcount] - maxcode - 1

			if code < extras {
				wps.w.pend_data |= (code << wps.w.pend_count)
				wps.w.pend_count += uint(bitcount - 1)
			} else {
				wps.w.pend_data |= (((code + extras) >> 1) << wps.w.pend_count)
				wps.w.pend_count += uint(bitcount - 1)
				wps.w.pend_data |= (((code + extras) & 1) << wps.w.pend_count)
				wps.w.pend_count++
			}
		}

		mid = uint(value)
	} else {
		for (high - low) > uint(wps.w.error_limit[channel]) {
			if uint(value) < mid {
				high = mid - 1
				mid = (high + low + 1) >> 1
				wps.w.pend_count++
			} else {
				low = mid
				mid = (high + low + 1) >> 1
				wps.w.pend_data |= bitset[wps.w.pend_count]
				wps.w.pend_count++
			}
		}
	}

	wps.w.pend_data |= uint(sign << wps.w.pend_count)
	wps.w.pend_count++

	if wps.w.holding_zero == 0 {
		flush_word(wps)
	}

	if (wps.wvcbits.active != 0) && (wps.w.error_limit[channel] != 0) {
		var code uint = uint(value) - low
		var maxcode uint = high - low
		var bitcount int = count_bits(maxcode)
		var extras uint = bitset[bitcount] - maxcode - 1

		if bitcount != 0 {
			if code < extras {
				putbits_correction(code, uint(bitcount-1), wps)
			} else {
				putbits_correction((code+extras)>>1, uint(bitcount-1), wps)
				putbit_correction((code+extras)&1, wps)
			}
		}
	}

	if (wps.wphdr.flags & HYBRID_BITRATE) != 0 {
		wps.w.slow_level[channel] -= ((wps.w.slow_level[channel] + SLO) >> SLS)
		wps.w.slow_level[channel] += mylog2(int(mid))
	}

	if sign == 1 {
		return int(^mid)
	} else {
		return (int)(mid)
	}
	return 0 // cannot get here, but keeps compiler happy
}

// This function is an optimized version of send_word() that only handles
// lossless (error_limit == 0). It does not return a value because it always
// encodes the exact value passed.
func send_word_lossless(wps *WavpackStream, value int, channel int) {
	var sign int
	if value < 0 {
		sign = 1
	} else {
		sign = 0
	}

	var ones_count uint
	var low uint
	var high uint

	if ((wps.w.median[0][0] & ^1) == 0) && (wps.w.holding_zero == 0) &&
		((wps.w.median[0][1] & ^1) == 0) {
		if wps.w.zeros_acc != 0 {
			if value != 0 {
				flush_word(wps)
			} else {
				wps.w.zeros_acc++

				return
			}
		} else if value != 0 {
			putbit_0(wps)
		} else {
			wps.w.median[0][0] = 0
			wps.w.median[1][0] = 0
			wps.w.median[2][0] = 0
			wps.w.median[0][1] = 0
			wps.w.median[1][1] = 0
			wps.w.median[2][1] = 0
			wps.w.zeros_acc = 1

			return
		}
	}

	if sign != 0 {
		value = ^value
	}

	if value < GET_MED(wps, 0, channel) {
		low = 0
		ones_count = 0
		high = uint(GET_MED(wps, 0, channel) - 1)
		DEC_MED0(wps, channel)
	} else {
		low = uint(GET_MED(wps, 0, channel))
		INC_MED0(wps, channel)

		if (value - int(low)) < GET_MED(wps, 1, channel) {
			ones_count = 1
			high = uint((int(low) + GET_MED(wps, 1, channel)) - 1)
			DEC_MED1(wps, channel)
		} else {
			low += uint(GET_MED(wps, 1, channel))
			INC_MED1(wps, channel)

			if (value - int(low)) < GET_MED(wps, 2, channel) {
				ones_count = 2
				high = uint((int(low) + GET_MED(wps, 2, channel)) - 1)
				DEC_MED2(wps, channel)
			} else {
				ones_count = uint(2 + ((value - int(low)) / GET_MED(wps, 2, channel)))
				low += uint(int(ones_count-2) * GET_MED(wps, 2, channel))
				high = uint((int(low) + GET_MED(wps, 2, channel)) - 1)
				INC_MED2(wps, channel)
			}
		}
	}

	if wps.w.holding_zero != 0 {
		if ones_count != 0 {
			wps.w.holding_one++
		}

		flush_word(wps)

		if ones_count != 0 {
			wps.w.holding_zero = 1
			ones_count--
		} else {
			wps.w.holding_zero = 0
		}
	} else {
		wps.w.holding_zero = 1
	}

	wps.w.holding_one = ones_count * 2

	if high != low {
		var maxcode uint = high - low
		var code uint = uint(value - int(low))
		var bitcount int = count_bits(maxcode)
		var extras uint = bitset[bitcount] - maxcode - 1

		if code < extras {
			wps.w.pend_data |= (code << wps.w.pend_count)
			wps.w.pend_count += uint(bitcount - 1)
		} else {
			wps.w.pend_data |= (((code + extras) >> 1) << wps.w.pend_count)
			wps.w.pend_count += uint(bitcount - 1)
			wps.w.pend_data |= (((code + extras) & 1) << wps.w.pend_count)
			wps.w.pend_count++
		}
	}

	wps.w.pend_data |= uint(sign << wps.w.pend_count)
	wps.w.pend_count++

	if wps.w.holding_zero == 0 {
		flush_word(wps)
	}
}

func putbit_0(wps *WavpackStream) {
	var bs Bitstream = wps.wvbits

	bs.bc++
	if bs.bc == 8 {
		wps.blockbuff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		(bs).sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvbits = bs
}

func putbit_1(wps *WavpackStream) {

	var bs Bitstream = wps.wvbits
	(bs).sr |= (1 << (bs).bc)

	bs.bc++
	if bs.bc == 8 {
		wps.blockbuff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		(bs).sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvbits = bs
}

func putbit(bit uint, wps *WavpackStream) {
	var bs Bitstream = wps.wvbits

	if bit != 0 {
		(bs).sr |= (1 << (bs).bc)
	}

	bs.bc++
	if bs.bc == 8 {
		wps.blockbuff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		(bs).sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvbits = bs
}

func putbits(value uint, nbits uint, wps *WavpackStream) {
	var bs Bitstream = wps.wvbits

	(bs).sr |= ((value) << (bs).bc)

	bs.bc += nbits
	if bs.bc >= 8 {
		for bs.bc >= 8 {
			wps.blockbuff[bs.buf_index] = byte(bs.sr)
			bs.buf_index++
			(bs).sr >>= 8

			bs.bc -= 8
			if bs.bc > 24 {
				(bs).sr |= ((value) >> ((nbits) - (bs).bc))
			}

			if bs.buf_index >= bs.end {
				bs_wrap(&bs) // error
			}
		}
	}
	wps.wvbits = bs
}

/* Bitstream routines for the correction file bits */

func putbit_correction_0(wps *WavpackStream) {
	var bs Bitstream = wps.wvcbits

	bs.bc++
	if bs.bc == 8 {
		wps.block2buff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		bs.sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvcbits = bs
}

func putbit_correction_1(wps *WavpackStream) {
	var bs Bitstream = wps.wvcbits
	(bs).sr |= (1 << (bs).bc)

	bs.bc++
	if bs.bc == 8 {
		wps.block2buff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		bs.sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvcbits = bs
}

func putbit_correction(bit uint, wps *WavpackStream) {
	var bs Bitstream = wps.wvcbits

	if bit != 0 {
		(bs).sr |= (1 << (bs).bc)
	}

	bs.bc++
	if bs.bc == 8 {
		wps.block2buff[bs.buf_index] = byte(bs.sr)
		bs.buf_index++
		bs.bc = 0
		bs.sr = 0

		if bs.buf_index >= bs.end {
			bs_wrap(&bs) // error
		}
	}
	wps.wvcbits = bs
}

func putbits_correction(value uint, nbits uint, wps *WavpackStream) {
	var bs Bitstream = wps.wvcbits
	(bs).sr |= ((value) << (bs).bc)

	bs.bc += nbits
	if bs.bc >= 8 {
		for bs.bc >= 8 {
			wps.block2buff[bs.buf_index] = byte(bs.sr)
			bs.buf_index++
			(bs).sr >>= 8

			bs.bc -= 8
			if bs.bc > 24 {
				(bs).sr |= ((value) >> ((nbits) - (bs).bc))
			}

			if bs.buf_index >= bs.end {
				bs_wrap(&bs) // error
			}
		}
	}
	wps.wvcbits = bs
}


// Used by send_word() and send_word_lossless() to actually send most the
// accumulated data onto the bitstream. This is also called directly from
// clients when all words have been sent.
func flush_word(wps *WavpackStream) {

	if wps.w.zeros_acc != 0 {
		var cbits int = count_bits(wps.w.zeros_acc)

		for cbits > 0 {
			putbit_1(wps)
			cbits--
		}

		putbit_0(wps)

		for wps.w.zeros_acc > 1 {
			putbit(wps.w.zeros_acc&1, wps)
			wps.w.zeros_acc >>= 1
		}

		wps.w.zeros_acc = 0
	}

	if wps.w.holding_one != 0 {
		if wps.w.holding_one >= LIMIT_ONES {
			var cbits int

			putbits((1<<LIMIT_ONES)-1, LIMIT_ONES+1, wps)
			wps.w.holding_one -= LIMIT_ONES
			cbits = count_bits(wps.w.holding_one)

			for cbits > 0 {
				putbit_1(wps)
				cbits--
			}

			putbit_0(wps)

			for wps.w.holding_one > 1 {
				putbit(wps.w.holding_one&1, wps)
				wps.w.holding_one >>= 1
			}

			wps.w.holding_zero = 0
		} else {
			putbits(bitmask[(int)(wps.w.holding_one)], wps.w.holding_one, wps)
		}

		wps.w.holding_one = 0
	}

	if wps.w.holding_zero != 0 {
		putbit_0(wps)
		wps.w.holding_zero = 0
	}

	if wps.w.pend_count != 0 {
		putbits(wps.w.pend_data, wps.w.pend_count, wps)
		wps.w.pend_count = 0
		wps.w.pend_data = 0
	}
}

// The concept of a base 2 logarithm is used in many parts of WavPack. It is
// a way of sufficiently accurately representing 32-bit signed and unsigned
// values storing only 16 bits (actually fewer). It is also used in the hybrid
// mode for quickly comparing the relative magnitude of large values (i.e.
// division) and providing smooth exponentials using only addition.
// These are not strict logarithms in that they become linear around zero and
// can therefore represent both zero and negative values. They have 8 bits
// of precision and in "roundtrip" conversions the total error never exceeds 1
// part in 225 except for the cases of +/-115 and +/-195 (which error by 1).
// This function returns the log2 for the specified 32-bit unsigned value.
// The maximum value allowed is about 0xff800000 and returns 8447.
func mylog2(avalue int) int {
	var dbits int

	avalue += (avalue >> 9)
	if (avalue) < (1 << 8) {
		dbits = nbits_table[avalue]

		return (dbits << 8) + log2_table[(int)(avalue<<uint(9-dbits))&0xff]
	} else {
		if avalue < (1 << 16) {
			dbits = nbits_table[(int)(avalue>>8)] + 8
		} else if avalue < (1 << 24) {
			dbits = nbits_table[(int)(avalue>>16)] + 16
		} else {
			dbits = nbits_table[(int)(avalue>>24)] + 24
		}

		return (dbits << 8) + log2_table[(avalue>>uint(dbits-9))&0xff]
	}
	return 0 // cannot actually reach this line, but keeps the compiler happy :)
}

// This function returns the log2 for the specified 32-bit signed value.
// All input values are valid and the return values are in the range of
// +/- 8192.
func log2s(value int) int {
	if value < 0 {
		return -mylog2(-value)
	} else {
		return mylog2(value)
	}
	return 0 // cannot actually reach this line, but keeps the compiler happy :)
}

// This function returns the original integer represented by the supplied
// logarithm (at least within the provided accuracy). The log is signed,
// but since a full 32-bit value is returned this can be used for unsigned
// conversions as well (i.e. the input range is -8192 to +8447).
func exp2s(log int) int {
	var value uint

	if log < 0 {
		return -exp2s(-log)
	}

	value = uint(exp2_table[log&0xff] | 0x100)

	log >>= 8
	if log <= 9 {
		return int((value >> uint(9-log)))
	} else {
		return int((value << uint(log-9)))
	}

	return 0 // again, cannot actually reach this line, but keeps the compiler happy :)
}


// These two functions convert internal weights (which are normally +/-1024)
// to and from an 8-bit signed character version for storage in metadata. The
// weights are clipped here in the case that they are outside that range.
func store_weight(weight int) byte {
	if weight > 1024 {
		weight = 1024
	} else if weight < -1024 {
		weight = -1024
	}

	if weight > 0 {
		weight -= ((weight + 64) >> 7)
	}

	return byte((weight + 4) >> 3)
}

func restore_weight(wt byte) int {
	var result int
	var weight int

	// we pass the value in as a byte (unsigned), but we actually need a signed value
	if wt > 127 {
		weight = int(wt) - 256
	} else {
		weight = int(wt)
	}

	result = weight << 3
	if result > 0 {
		result += ((result + 64) >> 7)
	}

	return result
}
