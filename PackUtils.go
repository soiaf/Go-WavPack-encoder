package wvencode

/*
** PackUtils.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

//////////////////////////////// local tables ///////////////////////////////
// These two tables specify the characteristics of the decorrelation filters.
// Each term represents one layer of the sequential filter, where positive
// values indicate the relative sample involved from the same channel (1=prev),
// 17 & 18 are special functions using the previous 2 samples, and negative
// values indicate cross channel decorrelation (in stereo only).
var very_high_terms = [17]int{18, 18, 2, 3, -2, 18, 2, 4, 7, 5, 3, 6, 8, -1, 18, 2, 0}
var high_terms = [11]int{18, 18, 18, -2, 2, 3, 5, -1, 17, 4, 0}
var default_terms = [6]int{18, 18, 2, 17, 3, 0}
var fast_terms = [3]int{18, 17, 0}


///////////////////////////// executable code ////////////////////////////////
// This function initializes everything required to pack WavPack bitstreams
// and must be called BEFORE any other function in this module.
func pack_init(wpc *WavpackContext) {
	var wps WavpackStream = wpc.stream
	var flags uint = wps.wphdr.flags
	var term_string []int
	var dpp_idx int = 0
	var ti int

	wps.sample_index = 0

	wps.dc.shaping_acc = make([]int, 2)   // initialise before first use
	wps.w.slow_level = make([]int, 2)     // initialise before first use
	wps.w.bitrate_delta = make([]int, 2)  // initialise before first use
	wps.dc.shaping_delta = make([]int, 2) // initialise before first use
	wps.dc.error = make([]int, 2)         // initialise before first use
	wps.w.error_limit = make([]int, 2)    // initialise before first use

	if (flags & HYBRID_SHAPE) > 0 {
		weight := wpc.config.Shaping_weight

		if weight <= -1000 {
			weight = -1000
		}

		wps.dc.shaping_acc[1] = weight << 16
		wps.dc.shaping_acc[0] = wps.dc.shaping_acc[1]
	}

	if (wpc.config.Flags & CONFIG_VERY_HIGH_FLAG) > 0 {
		term_string = very_high_terms[0:len(very_high_terms)]
	} else if (wpc.config.Flags & CONFIG_HIGH_FLAG) > 0 {
		term_string = high_terms[0:len(high_terms)]
	} else if (wpc.config.Flags & CONFIG_FAST_FLAG) > 0 {
		term_string = fast_terms[0:len(fast_terms)]
	} else {
		term_string = default_terms[0:len(default_terms)]
	}

	for ti = 0; ti < (len(term_string) - 1); ti++ {
		if (term_string[ti] >= 0) || ((flags & CROSS_DECORR) > 0) {
			wps.decorr_passes[dpp_idx].term = term_string[ti]

			wps.decorr_passes[dpp_idx].delta = 2
			dpp_idx++
		} else if (flags & MONO_FLAG) == 0 {
			wps.decorr_passes[dpp_idx].term = -3

			wps.decorr_passes[dpp_idx].delta = 2
			dpp_idx++
		}
	}

	wps.num_terms = dpp_idx

	wps.w.bitrate_acc = make([]uint, 2) // initialise before first use

	init_words(&wps)

	wpc.stream = wps
}


// Allocate room for and copy the decorrelation terms from the decorr_passes
// array into the specified metadata structure. Both the actual term id and
// the delta are packed into single characters.
func write_decorr_terms(wps WavpackStream, wpmd *WavpackMetadata) {
	var tcount int
	var byteptr []byte
	var byte_idx int = 0
	var dpp_idx int = 0

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data
	wpmd.id = ID_DECORR_TERMS

	for tcount = wps.num_terms; tcount > 0; dpp_idx++ {
		byteptr[byte_idx] = byte(((wps.decorr_passes[dpp_idx].term + 5) & 0x1f) |
			((wps.decorr_passes[dpp_idx].delta << 5) & 0xe0))

		byte_idx++
		tcount--
	}

	wpmd.byte_length = byte_idx
	wpmd.data = byteptr

}

// Allocate room for and copy the decorrelation term weights from the
// decorr_passes array into the specified metadata structure. The weights
// range +/-1024, but are rounded and truncated to fit in signed chars for
// metadata storage. Weights are separate for the two channels
func write_decorr_weights(wps *WavpackStream, wpmd *WavpackMetadata) {
	var tcount int
	var byteptr []byte
	var byte_idx int = 0
	var i int

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data
	wpmd.id = ID_DECORR_WEIGHTS

	for i = wps.num_terms - 1; i >= 0; i-- {
		if (store_weight(wps.decorr_passes[i].weight_A) != 0) ||
			(((wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0) &&
				(store_weight(wps.decorr_passes[i].weight_B) != 0)) {
			break
		}
	}

	tcount = i + 1

	for i = 0; i < wps.num_terms; i++ {
		if i < tcount {
			byteptr[byte_idx] = store_weight(wps.decorr_passes[i].weight_A)
			wps.decorr_passes[i].weight_A = restore_weight(byteptr[byte_idx])
			byte_idx++

			if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
				byteptr[byte_idx] = store_weight(wps.decorr_passes[i].weight_B)
				wps.decorr_passes[i].weight_B = restore_weight(byteptr[byte_idx])
				byte_idx++
			}
		} else {
			wps.decorr_passes[i].weight_B = 0
			wps.decorr_passes[i].weight_A = 0
		}

	}

	wpmd.byte_length = byte_idx
	wpmd.data = byteptr
}


// Allocate room for and copy the decorrelation samples from the decorr_passes
// array into the specified metadata structure. The samples are signed 32-bit
// values, but are converted to signed log2 values for storage in metadata.
// Values are stored for both channels and are specified from the first term
// with unspecified samples set to zero. The number of samples stored varies
// with the actual term value, so those must obviously be specified before
// these in the metadata list. Any number of terms can have their samples
// specified from no terms to all the terms, however I have found that
// sending more than the first term's samples is a waste. The "wcount"
// variable can be set to the number of terms to have their samples stored.
func write_decorr_samples(wps *WavpackStream, wpmd *WavpackMetadata) {
	var tcount int
	var wcount int = 1
	var temp int
	var byteptr []byte
	var byte_idx int = 0
	var dpp_idx int = 0

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data

	wpmd.id = ID_DECORR_SAMPLES

	for tcount = wps.num_terms; tcount > 0; tcount-- {
		if wcount != 0 {
			if wps.decorr_passes[dpp_idx].term > MAX_TERM {
				temp = log2s(wps.decorr_passes[dpp_idx].samples_A[0])
				wps.decorr_passes[dpp_idx].samples_A[0] = exp2s(temp)

				byteptr[byte_idx] = byte(temp)
				byte_idx++
				byteptr[byte_idx] = byte(temp >> 8)
				byte_idx++
				temp = log2s(wps.decorr_passes[dpp_idx].samples_A[1])
				wps.decorr_passes[dpp_idx].samples_A[1] = exp2s(temp)

				byteptr[byte_idx] = byte(temp)
				byte_idx++
				byteptr[byte_idx] = byte(temp >> 8)
				byte_idx++

				if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
					temp = log2s(wps.decorr_passes[dpp_idx].samples_B[0])
					wps.decorr_passes[dpp_idx].samples_B[0] = exp2s(temp)

					byteptr[byte_idx] = byte(temp)
					byte_idx++
					byteptr[byte_idx] = byte(temp >> 8)
					byte_idx++
					temp = log2s(wps.decorr_passes[dpp_idx].samples_B[1])
					wps.decorr_passes[dpp_idx].samples_B[1] = exp2s(temp)

					byteptr[byte_idx] = byte(temp)
					byte_idx++
					byteptr[byte_idx] = byte(temp >> 8)
					byte_idx++
				}
			} else if wps.decorr_passes[dpp_idx].term < 0 {
				temp = log2s(wps.decorr_passes[dpp_idx].samples_A[0])
				wps.decorr_passes[dpp_idx].samples_A[0] = exp2s(temp)

				byteptr[byte_idx] = byte(temp)
				byte_idx++
				byteptr[byte_idx] = byte(temp >> 8)
				byte_idx++
				temp = log2s(wps.decorr_passes[dpp_idx].samples_B[0])
				wps.decorr_passes[dpp_idx].samples_B[0] = exp2s(temp)

				byteptr[byte_idx] = byte(temp)
				byte_idx++
				byteptr[byte_idx] = byte(temp >> 8)
				byte_idx++
			} else {
				m := 0
				var cnt int = wps.decorr_passes[dpp_idx].term

				for cnt > 0 {
					temp = log2s(wps.decorr_passes[dpp_idx].samples_A[m])
					wps.decorr_passes[dpp_idx].samples_A[m] = exp2s(temp)

					byteptr[byte_idx] = byte(temp)
					byte_idx++
					byteptr[byte_idx] = byte(temp >> 8)
					byte_idx++

					if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
						temp = log2s(wps.decorr_passes[dpp_idx].samples_B[m])
						wps.decorr_passes[dpp_idx].samples_B[m] = exp2s(temp)

						byteptr[byte_idx] = byte(temp)
						byte_idx++
						byteptr[byte_idx] = byte(temp >> 8)
						byte_idx++
					}

					m++
					cnt--
				}
			}

			wcount--
		} else {
			for internalc := 0; internalc < MAX_TERM; internalc++ {
				wps.decorr_passes[dpp_idx].samples_A[internalc] = 0
				wps.decorr_passes[dpp_idx].samples_B[internalc] = 0
			}
		}

		dpp_idx++
	}

	wpmd.byte_length = byte_idx
	wpmd.data = byteptr
}

// Allocate room for and copy the noise shaping info into the specified
// metadata structure. These would normally be written to the
// "correction" file and are used for lossless reconstruction of
// hybrid data. The "delta" parameter is not yet used in encoding as it
// will be part of the "quality" mode.
func write_shaping_info(wps *WavpackStream, wpmd *WavpackMetadata) {
	var byteptr []byte
	var byte_idx int = 0
	var temp int

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data
	wpmd.id = ID_SHAPING_WEIGHTS

	temp = log2s(wps.dc.error[0])
	wps.dc.error[0] = exp2s(temp)
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++
	temp = log2s(wps.dc.shaping_acc[0])
	wps.dc.shaping_acc[0] = exp2s(temp)
	byteptr[byte_idx] = byte(temp)
	byte_idx++
	byteptr[byte_idx] = byte(temp >> 8)
	byte_idx++

	if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
		temp = log2s(wps.dc.error[1])
		wps.dc.error[1] = exp2s(temp)
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
		temp = log2s(wps.dc.shaping_acc[1])
		wps.dc.shaping_acc[1] = exp2s(temp)
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++
	}

	if (wps.dc.shaping_delta[0] | wps.dc.shaping_delta[1]) != 0 {
		temp = log2s(wps.dc.shaping_delta[0])
		wps.dc.shaping_delta[0] = exp2s(temp)
		byteptr[byte_idx] = byte(temp)
		byte_idx++
		byteptr[byte_idx] = byte(temp >> 8)
		byte_idx++

		if (wps.wphdr.flags & (MONO_FLAG | FALSE_STEREO)) == 0 {
			temp = log2s(wps.dc.shaping_delta[1])
			wps.dc.shaping_delta[1] = exp2s(temp)
			byteptr[byte_idx] = byte(temp)
			byte_idx++
			byteptr[byte_idx] = byte(temp >> 8)
			byte_idx++
		}
	}

	wpmd.byte_length = byte_idx
	wpmd.data = byteptr
}

// Allocate room for and copy the configuration information into the specified
// metadata structure. Currently, we just store the upper 3 bytes of
// config.flags and only in the first block of audio data. Note that this is
// for informational purposes not required for playback or decoding (like
// whether high or fast mode was specified).
func write_config_info(wpc *WavpackContext, wpmd *WavpackMetadata) {
	var byteptr []byte
	var byte_idx int = 0

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data

	wpmd.id = ID_CONFIG_BLOCK

	byteptr[byte_idx] = byte(wpc.config.Flags >> 8)
	byte_idx++
	byteptr[byte_idx] = byte(wpc.config.Flags >> 16)
	byte_idx++
	byteptr[byte_idx] = byte(wpc.config.Flags >> 24)
	byte_idx++

	wpmd.byte_length = byte_idx
	wpmd.data = byteptr
}

// Allocate room for and copy the non-standard sampling rateinto the specified
// metadata structure. We just store the lower 3 bytes of the sampling rate.
// Note that this would only be used when the sampling rate was not included
// in the table of 15 "standard" values.
func write_sample_rate(wpc *WavpackContext, wpmd *WavpackMetadata) {
	var byteptr []byte
	var byte_idx int = 0

	wpmd.data = wpmd.temp_data[0:len(wpmd.temp_data)]
	byteptr = wpmd.data

	wpmd.id = ID_SAMPLE_RATE
	byteptr[byte_idx] = byte(wpc.config.Sample_rate)
	byte_idx++
	byteptr[byte_idx] = byte(wpc.config.Sample_rate >> 8)
	byte_idx++
	byteptr[byte_idx] = byte(wpc.config.Sample_rate >> 16)
	byte_idx++
	wpmd.byte_length = byte_idx
	wpmd.data = byteptr
}


func pack_start_block(wpc *WavpackContext) int {
	var wps WavpackStream = wpc.stream
	var flags = wps.wphdr.flags
	var wpmd WavpackMetadata
	var chunkSize uint
	var copyRetVal int

	wps.lossy_block = FALSE
	wps.wphdr.crc = 0xffffffff
	wps.wphdr.block_samples = 0
	wps.wphdr.ckSize = WAVPACK_HEADER_SIZE - 8

	wps.blockbuff = make([]byte, BIT_BUFFER_SIZE)
	wps.block2buff = make([]byte, BIT_BUFFER_SIZE)

	wps.blockbuff[0] = byte(wps.wphdr.ckID[0])
	wps.blockbuff[1] = byte(wps.wphdr.ckID[1])
	wps.blockbuff[2] = byte(wps.wphdr.ckID[2])
	wps.blockbuff[3] = byte(wps.wphdr.ckID[3])
	wps.blockbuff[4] = byte(wps.wphdr.ckSize)
	wps.blockbuff[5] = byte(wps.wphdr.ckSize >> 8)
	wps.blockbuff[6] = byte(wps.wphdr.ckSize >> 16)
	wps.blockbuff[7] = byte(wps.wphdr.ckSize >> 24)
	wps.blockbuff[8] = byte(wps.wphdr.version)
	wps.blockbuff[9] = byte(wps.wphdr.version >> 8)
	wps.blockbuff[10] = byte(wps.wphdr.track_no)
	wps.blockbuff[11] = byte(wps.wphdr.index_no)
	wps.blockbuff[12] = byte(wps.wphdr.total_samples)
	wps.blockbuff[13] = byte(wps.wphdr.total_samples >> 8)
	wps.blockbuff[14] = byte(wps.wphdr.total_samples >> 16)
	wps.blockbuff[15] = byte(wps.wphdr.total_samples >> 24)
	wps.blockbuff[16] = byte(wps.wphdr.block_index)
	wps.blockbuff[17] = byte(wps.wphdr.block_index >> 8)
	wps.blockbuff[18] = byte(wps.wphdr.block_index >> 16)
	wps.blockbuff[19] = byte(wps.wphdr.block_index >> 24)
	wps.blockbuff[20] = byte(wps.wphdr.block_samples)
	wps.blockbuff[21] = byte(wps.wphdr.block_samples >> 8)
	wps.blockbuff[22] = byte(wps.wphdr.block_samples >> 16)
	wps.blockbuff[23] = byte(wps.wphdr.block_samples >> 24)
	wps.blockbuff[24] = byte(wps.wphdr.flags)
	wps.blockbuff[25] = byte(wps.wphdr.flags >> 8)
	wps.blockbuff[26] = byte(wps.wphdr.flags >> 16)
	wps.blockbuff[27] = byte(wps.wphdr.flags >> 24)
	wps.blockbuff[28] = byte(wps.wphdr.crc)
	wps.blockbuff[29] = byte(wps.wphdr.crc >> 8)
	wps.blockbuff[30] = byte(wps.wphdr.crc >> 16)
	wps.blockbuff[31] = byte(wps.wphdr.crc >> 24)

	write_decorr_terms(wps, &wpmd)
	copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

	if copyRetVal == FALSE {
		return FALSE
	}

	write_decorr_weights(&wps, &wpmd)
	copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

	if copyRetVal == FALSE {
		return FALSE
	}

	write_decorr_samples(&wps, &wpmd)
	copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

	if copyRetVal == FALSE {
		return FALSE
	}

	write_entropy_vars(&wps, &wpmd)
	copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

	if copyRetVal == FALSE {
		return FALSE
	}

	if ((flags & SRATE_MASK) == SRATE_MASK) &&
		(wpc.config.Sample_rate != 44100) {
		write_sample_rate(wpc, &wpmd)
		copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

		if copyRetVal == FALSE {
			return FALSE
		}

	}

	if (flags & HYBRID_FLAG) != 0 {
		write_hybrid_profile(&wps, &wpmd)
		copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

		if copyRetVal == FALSE {
			return FALSE
		}
	}

	if ((flags & INITIAL_BLOCK) > 0) && (wps.sample_index == 0) {
		write_config_info(wpc, &wpmd)
		copyRetVal, wps.blockbuff = copy_metadata(wpmd, wps.blockbuff, wps.blockend)

		if copyRetVal == FALSE {
			return FALSE
		}
	}

	chunkSize = uint((int(wps.blockbuff[4]) & 0xff) + ((int(wps.blockbuff[5]) & 0xff) << 8) +
		((int(wps.blockbuff[6]) & 0xff) << 16) + ((int(wps.blockbuff[7]) & 0xff) << 24))

	bs_open_write(&wps.wvbits, int(chunkSize+12), wps.blockend)

	if wpc.wvc_flag != 0 {
		wps.block2buff[0] = byte(wps.wphdr.ckID[0])
		wps.block2buff[1] = byte(wps.wphdr.ckID[1])
		wps.block2buff[2] = byte(wps.wphdr.ckID[2])
		wps.block2buff[3] = byte(wps.wphdr.ckID[3])
		wps.block2buff[4] = byte(wps.wphdr.ckSize)
		wps.block2buff[5] = byte(wps.wphdr.ckSize >> 8)
		wps.block2buff[6] = byte(wps.wphdr.ckSize >> 16)
		wps.block2buff[7] = byte(wps.wphdr.ckSize >> 24)
		wps.block2buff[8] = byte(wps.wphdr.version)
		wps.block2buff[9] = byte(wps.wphdr.version >> 8)
		wps.block2buff[10] = byte(wps.wphdr.track_no)
		wps.block2buff[11] = byte(wps.wphdr.index_no)
		wps.block2buff[12] = byte(wps.wphdr.total_samples)
		wps.block2buff[13] = byte(wps.wphdr.total_samples >> 8)
		wps.block2buff[14] = byte(wps.wphdr.total_samples >> 16)
		wps.block2buff[15] = byte(wps.wphdr.total_samples >> 24)
		wps.block2buff[16] = byte(wps.wphdr.block_index)
		wps.block2buff[17] = byte(wps.wphdr.block_index >> 8)
		wps.block2buff[18] = byte(wps.wphdr.block_index >> 16)
		wps.block2buff[19] = byte(wps.wphdr.block_index >> 24)
		wps.block2buff[20] = byte(wps.wphdr.block_samples)
		wps.block2buff[21] = byte(wps.wphdr.block_samples >> 8)
		wps.block2buff[22] = byte(wps.wphdr.block_samples >> 16)
		wps.block2buff[23] = byte(wps.wphdr.block_samples >> 24)
		wps.block2buff[24] = byte(wps.wphdr.flags)
		wps.block2buff[25] = byte(wps.wphdr.flags >> 8)
		wps.block2buff[26] = byte(wps.wphdr.flags >> 16)
		wps.block2buff[27] = byte(wps.wphdr.flags >> 24)
		wps.block2buff[28] = byte(wps.wphdr.crc)
		wps.block2buff[29] = byte(wps.wphdr.crc >> 8)
		wps.block2buff[30] = byte(wps.wphdr.crc >> 16)
		wps.block2buff[31] = byte(wps.wphdr.crc >> 24)

		if (flags & HYBRID_SHAPE) != 0 {
			write_shaping_info(&wps, &wpmd)
			copyRetVal, wps.block2buff = copy_metadata(wpmd, wps.block2buff, wps.block2end)

			if copyRetVal == FALSE {
				return FALSE
			}
		}

		chunkSize = uint((int(wps.block2buff[4]) & 0xff) + ((int(wps.block2buff[5]) & 0xff) << 8) +
			((int(wps.block2buff[6]) & 0xff) << 16) + ((int(wps.block2buff[7]) & 0xff) << 24))

		bs_open_write(&wps.wvcbits, (int)(chunkSize+12), wps.block2end)
	} else {
		wps.block2buff[0] = 0
	}

	wpc.stream = wps

	return TRUE
}


// Pack the given samples into the block currently being assembled. This function
// checks the available space each sample so that it can return prematurely to
// indicate that the blocks must be terminated. The return value is the number
// of actual samples packed and will be the same as the provided sample_count
// in no error occurs.
func pack_samples(wpc *WavpackContext, buffer []int, sample_count uint) uint {
	var wps WavpackStream = wpc.stream

	var flags = wps.wphdr.flags

	var tcount int
	var lossy int = 0
	var m int
	var byte_idx int = 0
	var dpp_idx int = 0
	var crc int
	var crc2 int
	var i uint
	var bptr []int
	var block_samples uint

	if sample_count == 0 {
		return 0
	}

	byte_idx = wpc.Byte_idx // Get the index position for the buffer holding the input WAV data

	i = 0

	block_samples = uint((int(wps.blockbuff[23]) & 0xFF) << 24)
	block_samples += uint((int(wps.blockbuff[22]) & 0xFF) << 16)
	block_samples += uint((int(wps.blockbuff[21]) & 0xFF) << 8)
	block_samples += uint(int(wps.blockbuff[20]) & 0XFF)
	m = (int(block_samples) & (MAX_TERM - 1))

	crc = int((int(wps.blockbuff[31]) & 0xFF) << 24)
	crc += int((int(wps.blockbuff[30]) & 0xFF) << 16)
	crc += int((int(wps.blockbuff[29]) & 0xFF) << 8)
	crc += int(int(wps.blockbuff[28]) & 0xFF)

	crc2 = 0

	if wpc.wvc_flag != 0 {
		crc2 = int(int(wps.block2buff[31]&0xFF) << 24)
		crc2 += int(int(wps.block2buff[30]&0xFF) << 16)
		crc2 += int(int(wps.block2buff[29]&0xFF) << 8)
		crc2 += int(int(wps.block2buff[28]) & 0xFF)
	}

	/////////////////////// handle lossless mono mode /////////////////////////
	if ((flags & HYBRID_FLAG) == 0) &&
		((flags & (MONO_FLAG | FALSE_STEREO)) != 0) {

		bptr = buffer
		for i = 0; i < sample_count; i++ {
			var code uint

			if bs_remain_write(wps.wvbits) < 64 {
				break
			}

			code = uint(bptr[byte_idx])
			crc = (crc * 3) + int(code)
			byte_idx++

			dpp_idx = 0

			for tcount = wps.num_terms; tcount > 0; tcount-- {
				var sam int

				if wps.decorr_passes[dpp_idx].term > MAX_TERM {
					if (wps.decorr_passes[dpp_idx].term & 1) != 0 {
						sam = (2 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]
					} else {
						sam = ((3 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]) >> 1
					}

					wps.decorr_passes[dpp_idx].samples_A[1] = wps.decorr_passes[dpp_idx].samples_A[0]
					wps.decorr_passes[dpp_idx].samples_A[0] = int(code)
				} else {
					sam = wps.decorr_passes[dpp_idx].samples_A[m]

					wps.decorr_passes[dpp_idx].samples_A[(m+wps.decorr_passes[dpp_idx].term)&
						(MAX_TERM-1)] = int(code)
				}

				code -= uint(apply_weight(wps.decorr_passes[dpp_idx].weight_A, sam))
				wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
					wps.decorr_passes[dpp_idx].delta, sam, int(code))

				dpp_idx++
			}

			m = (m + 1) & (MAX_TERM - 1)

			send_word_lossless(&wps, int(code), 0)
		}

		wpc.Byte_idx = byte_idx

		//////////////////// handle the lossless stereo mode //////////////////////
	} else if ((flags & HYBRID_FLAG) == 0) &&
		((flags & (MONO_FLAG | FALSE_STEREO)) == 0) {
		bptr = buffer
		for i = 0; i < sample_count; i++ {
			var left int
			var right int
			var sam_A int
			var sam_B int

			if bs_remain_write(wps.wvbits) < 128 {
				break
			}

			left = int(bptr[byte_idx])
			crc = (crc * 3) + int(left)
			right = int(bptr[byte_idx+1])
			crc = (crc * 3) + int(right)

			if (flags & JOINT_STEREO) > 0 {
				left -= right
				right += (left >> 1)
			}

			dpp_idx = 0

			for tcount = wps.num_terms; tcount > 0; tcount-- {
				if wps.decorr_passes[dpp_idx].term > 0 {
					if wps.decorr_passes[dpp_idx].term > MAX_TERM {
						if (wps.decorr_passes[dpp_idx].term & 1) != 0 {
							sam_A = (2 * wps.decorr_passes[dpp_idx].samples_A[0]) -
								wps.decorr_passes[dpp_idx].samples_A[1]
							sam_B = (2 * wps.decorr_passes[dpp_idx].samples_B[0]) -
								wps.decorr_passes[dpp_idx].samples_B[1]
						} else {
							sam_A = ((3 * wps.decorr_passes[dpp_idx].samples_A[0]) -
								wps.decorr_passes[dpp_idx].samples_A[1]) >> 1
							sam_B = ((3 * wps.decorr_passes[dpp_idx].samples_B[0]) -
								wps.decorr_passes[dpp_idx].samples_B[1]) >> 1
						}

						wps.decorr_passes[dpp_idx].samples_A[1] = wps.decorr_passes[dpp_idx].samples_A[0]
						wps.decorr_passes[dpp_idx].samples_B[1] = wps.decorr_passes[dpp_idx].samples_B[0]
						wps.decorr_passes[dpp_idx].samples_A[0] = left
						wps.decorr_passes[dpp_idx].samples_B[0] = right
					} else {
						var k int = (m + wps.decorr_passes[dpp_idx].term) & (MAX_TERM - 1)

						sam_A = wps.decorr_passes[dpp_idx].samples_A[m]
						sam_B = wps.decorr_passes[dpp_idx].samples_B[m]
						wps.decorr_passes[dpp_idx].samples_A[k] = left
						wps.decorr_passes[dpp_idx].samples_B[k] = right
					}

					left -= apply_weight(wps.decorr_passes[dpp_idx].weight_A, sam_A)
					right -= apply_weight(wps.decorr_passes[dpp_idx].weight_B, sam_B)
					wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta, sam_A, left)
					wps.decorr_passes[dpp_idx].weight_B = update_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].delta, sam_B, right)
				} else {
					if wps.decorr_passes[dpp_idx].term == -2 {
						sam_A = right
					} else {
						sam_A = wps.decorr_passes[dpp_idx].samples_A[0]
					}

					if wps.decorr_passes[dpp_idx].term == -1 {
						sam_B = left
					} else {
						sam_B = wps.decorr_passes[dpp_idx].samples_B[0]
					}

					wps.decorr_passes[dpp_idx].samples_A[0] = right
					wps.decorr_passes[dpp_idx].samples_B[0] = left
					left -= apply_weight(wps.decorr_passes[dpp_idx].weight_A, sam_A)
					right -= apply_weight(wps.decorr_passes[dpp_idx].weight_B, sam_B)
					wps.decorr_passes[dpp_idx].weight_A = update_weight_clip(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta, sam_A, uint(left))
					wps.decorr_passes[dpp_idx].weight_B = update_weight_clip(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].delta, sam_B, uint(right))
				}

				dpp_idx++
			}

			m = (m + 1) & (MAX_TERM - 1)
			send_word_lossless(&wps, left, 0)
			send_word_lossless(&wps, right, 1)

			byte_idx += 2
		}

		wpc.Byte_idx = byte_idx

		/////////////////// handle the lossy/hybrid mono mode /////////////////////
	} else if ((flags & HYBRID_FLAG) != 0) &&
		((flags & (MONO_FLAG | FALSE_STEREO)) != 0) {
		bptr = buffer
		for i = 0; i < sample_count; i++ {
			var code int
			var temp int

			if (bs_remain_write(wps.wvbits) < 64) ||
				((wpc.wvc_flag != 0) && (bs_remain_write(wps.wvcbits) < 64)) {
				break
			}

			code = int(bptr[byte_idx])
			crc2 = (crc2 * 3) + code
			byte_idx++

			if (flags & HYBRID_SHAPE) != 0 {
				wps.dc.shaping_acc[0] += wps.dc.shaping_delta[0]
				var shaping_weight int = (wps.dc.shaping_acc[0]) >> 16
				temp = -apply_weight(shaping_weight, wps.dc.error[0])

				if ((flags & NEW_SHAPING) != 0) && (shaping_weight < 0) &&
					(temp != 0) {
					if temp == wps.dc.error[0] {
						if temp < 0 {
							temp = temp + 1
						} else {
							temp = temp - 1
						}
					}

					wps.dc.error[0] = -code
					code += temp
				} else {
					code += temp
					wps.dc.error[0] = -(code)
				}
			}

			dpp_idx = 0

			for tcount = wps.num_terms; tcount > 0; tcount-- {
				if wps.decorr_passes[dpp_idx].term > MAX_TERM {
					if (wps.decorr_passes[dpp_idx].term & 1) != 0 {
						wps.decorr_passes[dpp_idx].samples_A[2] = (2 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]
					} else {
						wps.decorr_passes[dpp_idx].samples_A[2] = ((3 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]) >> 1
					}

					wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].samples_A[2])
					code -= (wps.decorr_passes[dpp_idx].aweight_A)
				} else {
					wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].samples_A[m])
					code -= (wps.decorr_passes[dpp_idx].aweight_A)
				}

				dpp_idx++
			}

			code = send_word(&wps, code, 0)

			dpp_idx--

			for dpp_idx >= 0 {
				if wps.decorr_passes[dpp_idx].term > MAX_TERM {
					wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_A[2], code)
					wps.decorr_passes[dpp_idx].samples_A[1] = wps.decorr_passes[dpp_idx].samples_A[0]
					code += wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].samples_A[0] = code
				} else {
					var sam int = wps.decorr_passes[dpp_idx].samples_A[m]

					wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta, sam, code)

					code += wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].samples_A[(m+wps.decorr_passes[dpp_idx].term)&
						(MAX_TERM-1)] = code
				}

				dpp_idx--
			}

			wps.dc.error[0] += code
			m = (m + 1) & (MAX_TERM - 1)

			crc = (crc * 3) + int(code)
			if crc != int(crc2) {
				lossy = TRUE
			}
		}

		wpc.Byte_idx = byte_idx

		/////////////////// handle the lossy/hybrid stereo mode ///////////////////
	} else if ((flags & HYBRID_FLAG) != 0) &&
		((flags & (MONO_FLAG | FALSE_STEREO)) == 0) {
		bptr = buffer
		for i = 0; i < sample_count; i++ {
			var left int
			var right int
			var temp int
			var shaping_weight int

			if (bs_remain_write(wps.wvbits) < 128) ||
				((wpc.wvc_flag != 0) && (bs_remain_write(wps.wvcbits) < 128)) {
				break
			}

			left = int(bptr[byte_idx])
			byte_idx++
			right = int(bptr[byte_idx])
			crc2 = (((crc2 * 3) + left) * 3) + right
			byte_idx++

			if (flags & HYBRID_SHAPE) != 0 {
				wps.dc.shaping_acc[0] += wps.dc.shaping_delta[0]
				shaping_weight = (wps.dc.shaping_acc[0]) >> 16
				temp = -apply_weight(shaping_weight, wps.dc.error[0])

				if ((flags & NEW_SHAPING) != 0) && (shaping_weight < 0) &&
					(temp != 0) {
					if temp == wps.dc.error[0] {
						if temp < 0 {
							temp = temp + 1
						} else {
							temp = temp - 1
						}
					}

					wps.dc.error[0] = -left
					left += temp
				} else {
					left += temp
					wps.dc.error[0] = -(left)
				}

				wps.dc.shaping_acc[1] += wps.dc.shaping_delta[1]
				shaping_weight = (wps.dc.shaping_acc[1]) >> 16
				temp = -apply_weight(shaping_weight, wps.dc.error[1])

				if ((flags & NEW_SHAPING) != 0) && (shaping_weight < 0) &&
					(temp != 0) {
					if temp == wps.dc.error[1] {
						if temp < 0 {
							temp = temp + 1
						} else {
							temp = temp - 1
						}
					}

					wps.dc.error[1] = -right
					right += temp
				} else {
					right += temp
					wps.dc.error[1] = -(right)
				}
			}

			if (flags & JOINT_STEREO) != 0 {
				left -= right
				right += (left >> 1)
			}

			dpp_idx = 0

			for tcount = wps.num_terms; tcount > 0; tcount-- {
				if wps.decorr_passes[dpp_idx].term > MAX_TERM {
					if (wps.decorr_passes[dpp_idx].term & 1) != 0 {
						wps.decorr_passes[dpp_idx].samples_A[2] = (2 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]
						wps.decorr_passes[dpp_idx].samples_B[2] = (2 * wps.decorr_passes[dpp_idx].samples_B[0]) -
							wps.decorr_passes[dpp_idx].samples_B[1]
					} else {
						wps.decorr_passes[dpp_idx].samples_A[2] = ((3 * wps.decorr_passes[dpp_idx].samples_A[0]) -
							wps.decorr_passes[dpp_idx].samples_A[1]) >> 1
						wps.decorr_passes[dpp_idx].samples_B[2] = ((3 * wps.decorr_passes[dpp_idx].samples_B[0]) -
							wps.decorr_passes[dpp_idx].samples_B[1]) >> 1
					}
					wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].samples_A[2])
					left -= (wps.decorr_passes[dpp_idx].aweight_A)
					wps.decorr_passes[dpp_idx].aweight_B = apply_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].samples_B[2])
					right -= (wps.decorr_passes[dpp_idx].aweight_B)
				} else if wps.decorr_passes[dpp_idx].term > 0 {
					wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].samples_A[m])
					left -= (wps.decorr_passes[dpp_idx].aweight_A)

					wps.decorr_passes[dpp_idx].aweight_B = apply_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].samples_B[m])
					right -= (wps.decorr_passes[dpp_idx].aweight_B)
				} else {
					if wps.decorr_passes[dpp_idx].term == -1 {
						wps.decorr_passes[dpp_idx].samples_B[0] = left
					} else if wps.decorr_passes[dpp_idx].term == -2 {
						wps.decorr_passes[dpp_idx].samples_A[0] = right
					}

					wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].samples_A[0])
					left -= wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].aweight_B = apply_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].samples_B[0])
					right -= wps.decorr_passes[dpp_idx].aweight_B
				}

				dpp_idx++
			}

			left = send_word(&wps, left, 0)
			right = send_word(&wps, right, 1)

			dpp_idx--

			for dpp_idx >= 0 {
				if wps.decorr_passes[dpp_idx].term > MAX_TERM {
					wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_A[2], left)
					wps.decorr_passes[dpp_idx].weight_B = update_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_B[2], right)

					wps.decorr_passes[dpp_idx].samples_A[1] = wps.decorr_passes[dpp_idx].samples_A[0]
					wps.decorr_passes[dpp_idx].samples_B[1] = wps.decorr_passes[dpp_idx].samples_B[0]

					left += wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].samples_A[0] = left
					right += wps.decorr_passes[dpp_idx].aweight_B
					wps.decorr_passes[dpp_idx].samples_B[0] = right
				} else if wps.decorr_passes[dpp_idx].term > 0 {
					var k int = (m + wps.decorr_passes[dpp_idx].term) & (MAX_TERM - 1)

					wps.decorr_passes[dpp_idx].weight_A = update_weight(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_A[m], left)
					left += wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].samples_A[k] = left

					wps.decorr_passes[dpp_idx].weight_B = update_weight(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_B[m], right)
					right += wps.decorr_passes[dpp_idx].aweight_B
					wps.decorr_passes[dpp_idx].samples_B[k] = right
				} else {
					if wps.decorr_passes[dpp_idx].term == -1 {
						wps.decorr_passes[dpp_idx].samples_B[0] = left +
							wps.decorr_passes[dpp_idx].aweight_A
						wps.decorr_passes[dpp_idx].aweight_B = apply_weight(wps.decorr_passes[dpp_idx].weight_B,
							wps.decorr_passes[dpp_idx].samples_B[0])
					} else if wps.decorr_passes[dpp_idx].term == -2 {
						wps.decorr_passes[dpp_idx].samples_A[0] = right +
							wps.decorr_passes[dpp_idx].aweight_B
						wps.decorr_passes[dpp_idx].aweight_A = apply_weight(wps.decorr_passes[dpp_idx].weight_A,
							wps.decorr_passes[dpp_idx].samples_A[0])
					}

					wps.decorr_passes[dpp_idx].weight_A = update_weight_clip(wps.decorr_passes[dpp_idx].weight_A,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_A[0], uint(left))
					wps.decorr_passes[dpp_idx].weight_B = update_weight_clip(wps.decorr_passes[dpp_idx].weight_B,
						wps.decorr_passes[dpp_idx].delta,
						wps.decorr_passes[dpp_idx].samples_B[0], uint(right))
					left += wps.decorr_passes[dpp_idx].aweight_A
					wps.decorr_passes[dpp_idx].samples_B[0] = left
					right += wps.decorr_passes[dpp_idx].aweight_B
					wps.decorr_passes[dpp_idx].samples_A[0] = right
				}

				dpp_idx--
			}

			if (flags & JOINT_STEREO) != 0 {
				right -= (left >> 1)
				left += right
			}

			wps.dc.error[0] += left
			wps.dc.error[1] += right
			m = (m + 1) & (MAX_TERM - 1)

			crc = (((crc * 3) + int(left)) * 3) + int(right)
			if crc != crc2 {
				lossy = TRUE
			}
		}

		wpc.Byte_idx = byte_idx
	}

	block_samples = uint((int(wps.blockbuff[23]) & 0xFF) << 24)
	block_samples += uint((int(wps.blockbuff[22]) & 0xFF) << 16)
	block_samples += uint((int(wps.blockbuff[21]) & 0xFF) << 8)
	block_samples += uint(int(wps.blockbuff[20]) & 0XFF)

	block_samples = block_samples + i

	wps.blockbuff[20] = byte(block_samples)
	wps.blockbuff[21] = byte(block_samples >> 8)
	wps.blockbuff[22] = byte(block_samples >> 16)
	wps.blockbuff[23] = byte(block_samples >> 24)

	wps.blockbuff[28] = byte(crc)
	wps.blockbuff[29] = byte(crc >> 8)
	wps.blockbuff[30] = byte(crc >> 16)
	wps.blockbuff[31] = byte(crc >> 24)

	if wpc.wvc_flag != 0 {
		block_samples = uint(int(wps.block2buff[23]&0xFF) << 24)
		block_samples += uint(int(wps.block2buff[22]&0xFF) << 16)
		block_samples += uint(int(wps.block2buff[21]&0xFF) << 8)
		block_samples += uint(int(wps.block2buff[20] & 0XFF))

		block_samples = block_samples + i

		wps.block2buff[20] = byte(block_samples)
		wps.block2buff[21] = byte(block_samples >> 8)
		wps.block2buff[22] = byte(block_samples >> 16)
		wps.block2buff[23] = byte(block_samples >> 24)

		wps.block2buff[28] = byte(crc2)
		wps.block2buff[29] = byte(crc2 >> 8)
		wps.block2buff[30] = byte(crc2 >> 16)
		wps.block2buff[31] = byte(crc2 >> 24)
	}

	if lossy != 0 {
		wps.lossy_block = TRUE
	}

	wps.sample_index += int(i)

	wpc.stream = wps

	return i
}

func apply_weight(weight int, sample int) int {
	return (((((sample & 0xffff) * weight) >> 9) + (((sample & ^0xffff) >> 9) * weight) + 1) >> 1)
}

func update_weight(weight int, delta int, source int, result int) int {
	if (source != 0) && (result != 0) {
		weight += ((((source ^ result) >> 30) | 1) * delta)
	}

	return weight
}

func update_weight_clip(weight int, delta int, source int, result uint) int {
	if source != 0 && result != 0 {
		if (source ^ int(result)) < 0 {
			weight -= delta
			if weight < -1024 {
				weight = -1024
			}
		} else {
			weight += delta
			if weight > 1024 {
				weight = 1024
			}
		}
	}

	return weight
}

// Once all the desired samples have been packed into the WavPack block being
// built, this function is called to prepare it for writing. Basically, this
// means just closing the bitstreams because the block_samples and crc fields
// of the WavpackHeader are updated during packing.
func pack_finish_block(wpc *WavpackContext) int {
	var wps WavpackStream = wpc.stream
	var lossy int = wps.lossy_block
	var tcount int
	var m int
	var data_count int
	var block_samples uint
	var chunkSize int = 0
	var dpp_idx int = 0

	block_samples = uint((int(wps.blockbuff[23]) & 0xFF) << 24)
	block_samples += uint((int(wps.blockbuff[22]) & 0xFF) << 16)
	block_samples += uint((int(wps.blockbuff[21]) & 0xFF) << 8)
	block_samples += uint(int(wps.blockbuff[20]) & 0XFF)

	m = (int(block_samples) & (MAX_TERM - 1))

	if m != 0 {
		for tcount = wps.num_terms; tcount > 0; dpp_idx++ {
			if (wps.decorr_passes[dpp_idx].term > 0) &&
				(wps.decorr_passes[dpp_idx].term <= MAX_TERM) {
				var temp_A []int // MAX_TERM
				var temp_B []int // MAX_TERM
				var k int

				temp_A = wps.decorr_passes[dpp_idx].samples_A[0:len(wps.decorr_passes[dpp_idx].samples_A)]
				temp_B = wps.decorr_passes[dpp_idx].samples_B[0:len(wps.decorr_passes[dpp_idx].samples_B)]

				for k = 0; k < MAX_TERM; k++ {
					wps.decorr_passes[dpp_idx].samples_A[k] = temp_A[m]
					wps.decorr_passes[dpp_idx].samples_B[k] = temp_B[m]
					m = (m + 1) & (MAX_TERM - 1)
				}
			}

			tcount--
		}
	}

	flush_word(&wps)

	data_count = bs_close_write(&wps)

	if data_count != 0 {
		if data_count != -1 {
			var cptr_idx uint = 0

			chunkSize = (int(wps.blockbuff[4]) & 0xff) + ((int(wps.blockbuff[5]) & 0xff) << 8) +
				((int(wps.blockbuff[6]) & 0xff) << 16) + ((int(wps.blockbuff[7]) & 0xff) << 24)

			cptr_idx = uint(chunkSize) + 8

			wps.blockbuff[cptr_idx] = byte(ID_WV_BITSTREAM | ID_LARGE)
			cptr_idx++
			wps.blockbuff[cptr_idx] = byte(data_count >> 1)
			cptr_idx++
			wps.blockbuff[cptr_idx] = byte(data_count >> 9)
			cptr_idx++
			wps.blockbuff[cptr_idx] = byte(data_count >> 17)

			chunkSize = chunkSize + data_count + 4

			wps.blockbuff[4] = byte(chunkSize)
			wps.blockbuff[5] = byte(chunkSize >> 8)
			wps.blockbuff[6] = byte(chunkSize >> 16)
			wps.blockbuff[7] = byte(chunkSize >> 24)

		} else {
			return FALSE
		}
	}

	if wpc.wvc_flag != 0 {
		data_count = bs_close_correction_write(&wps)

		if (data_count != 0) && (lossy != 0) {
			if data_count != -1 {
				var cptr_idx uint = 0
				chunkSize = int(int(wps.block2buff[4]&0xff) + (int(wps.block2buff[5]&0xff) << 8) +
					(int(wps.block2buff[6]&0xff) << 16) + (int(wps.block2buff[7]&0xff) << 24))

				cptr_idx = uint(chunkSize) + 8

				wps.block2buff[cptr_idx] = byte(ID_WVC_BITSTREAM | ID_LARGE)
				cptr_idx++
				wps.block2buff[cptr_idx] = byte(data_count >> 1)
				cptr_idx++
				wps.block2buff[cptr_idx] = byte(data_count >> 9)
				cptr_idx++
				wps.block2buff[cptr_idx] = byte(data_count >> 17)
				cptr_idx++

				chunkSize = chunkSize + data_count + 4
				wps.block2buff[4] = byte(chunkSize)
				wps.block2buff[5] = byte(chunkSize >> 8)
				wps.block2buff[6] = byte(chunkSize >> 16)
				wps.block2buff[7] = byte(chunkSize >> 24)
			} else {
				return FALSE
			}
		}
	} else if lossy != 0 {
		wpc.lossy_blocks = TRUE
	}

	wpc.stream = wps

	return TRUE
}


// Copy the specified metadata item to the WavPack block being contructed. This
// function tests for writing past the end of the available space, however the
// rest of the code is designed so that can't occur.
// Prepare a WavPack block for writing. The block will be written at
// "wps.blockbuff" and "wps.blockend" points to the end of the available
// space. If a wvc file is being written, then block2buff and block2end are
// also used. This also sets up the bitstreams so that pack_samples() can be
// called next with actual sample data. To find out how much data was written
// the caller must look at the ckSize field of the written WavpackHeader, NOT
// the one in the WavpackStream. A return value of FALSE indicates an error.

func copy_metadata(wpmd WavpackMetadata, buffer_start []byte, buffer_end int) (int, []byte) {
	var mdsize uint = uint(wpmd.byte_length + (wpmd.byte_length & 1))
	var chunkSize uint
	var bufIdx int = 0

	if (wpmd.byte_length & 1) != 0 {
		(wpmd.data)[wpmd.byte_length] = 0
	}

	if wpmd.byte_length > 510 {
		mdsize += 4
	} else {
		mdsize += 2
	}

	chunkSize = uint((int(buffer_start[4]) & 0xff) + ((int(buffer_start[5]) & 0xff) << 8) +
		((int(buffer_start[6]) & 0xff) << 16) + ((int(buffer_start[7]) & 0xff) << 24))

	bufIdx = (int)(chunkSize + 8)

	if (bufIdx + int(mdsize)) >= buffer_end {
		return FALSE, buffer_start
	}

	var oddsizecheck int = 0

	if (wpmd.byte_length & 1) != 0 {
		oddsizecheck = ID_ODD_SIZE
	} else {
		oddsizecheck = 0
	}

	buffer_start[bufIdx] = byte(wpmd.id | oddsizecheck)
	buffer_start[bufIdx+1] = byte((wpmd.byte_length + 1) >> 1)

	if wpmd.byte_length > 510 {
		buffer_start[bufIdx] |= byte(ID_LARGE)
		buffer_start[bufIdx+2] = byte((wpmd.byte_length + 1) >> 9)
		buffer_start[bufIdx+3] = byte((wpmd.byte_length + 1) >> 17)
	}

	if (len(wpmd.data) != 0) && (wpmd.byte_length != 0) {
		if wpmd.byte_length > 510 {
			buffer_start[bufIdx] |= byte(ID_LARGE)
			buffer_start[bufIdx+2] = byte((wpmd.byte_length + 1) >> 9)
			buffer_start[bufIdx+3] = byte((wpmd.byte_length + 1) >> 17)

			for t := 0; t < int(mdsize-4); t++ {
				buffer_start[t+(bufIdx+4)] = byte(wpmd.data[t])
			}

		} else {
			for t := 0; t < int(mdsize-2); t++ {
				buffer_start[t+(bufIdx+2)] = byte(wpmd.data[t])
			}
		}
	}

	chunkSize += mdsize

	buffer_start[4] = byte(chunkSize)
	buffer_start[5] = byte(chunkSize >> 8)
	buffer_start[6] = byte(chunkSize >> 16)
	buffer_start[7] = byte(chunkSize >> 24)

	return TRUE, buffer_start
}
