package main

/*
** WvEncode.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */


import (
	"fmt"
	"os"
	"math"
	"strconv"
	"./wvencode"
)

const usage0 = "\n"
const usage1 string = " Usage:   gowavpack [-options] infile.wav outfile.wv [outfile.wvc]\n"
const usage2 string = " (default is lossless)\n"
const usage3 string = "\n"
const usage4 string = "  Options: \n       -bn = enable hybrid compression, n = 2.0 to 16.0 bits/sample\n"
const usage5 string = "       -c  = create correction file (.wvc) for hybrid mode (=lossless)\n"
const usage6 string = "       -cc = maximum hybrid compression (hurts lossy quality & decode speed)\n"
const usage7 string = "       -f  = fast mode (fast, but some compromise in compression ratio)\n"
const usage8 string = "       -h  = high quality (better compression in all modes, but slower)\n"
const usage9 string = "       -hh = very high quality (best compression in all modes, but slowest\n"
const usage10 string = "                              and NOT recommended for portable hardware use)\n"
const usage11 string = "       -jn = joint-stereo override (0 = left/right, 1 = mid/side)\n"
const usage12 string = "       -sn = noise shaping override (hybrid only, n = -1.0 to 1.0, 0 = off)\n"


func usage() {
	fmt.Printf(usage0)
	fmt.Printf(usage1)
	fmt.Printf(usage2)
	fmt.Printf(usage3)
	fmt.Printf(usage4)
	fmt.Printf(usage5)
	fmt.Printf(usage6)
	fmt.Printf(usage7)
	fmt.Printf(usage8)
	fmt.Printf(usage9)
	fmt.Printf(usage10)
	fmt.Printf(usage11)
	fmt.Printf(usage12)

	os.Exit(1)
}

func main() {
	// This is the main module for the demonstration WavPack command-line
	// encoder using the "tiny encoder". It accepts a source WAV file, a
	// destination WavPack file (.wv) and an optional WavPack correction file
	// (.wvc) on the command-line. It supports all 4 encoding qualities in
	// pure lossless, hybrid lossy and hybrid lossless modes. Valid input are
	// mono or stereo integer WAV files with bitdepths from 8 to 24.
	// This program (and the tiny encoder) do not handle placing the WAV RIFF
	// header into the WavPack file. The latest version of the regular WavPack
	// unpacker (4.40) and the "tiny decoder" will generate the RIFF header
	// automatically on unpacking. However, older versions of the command-line
	// program will complain about this and require unpacking in "raw" mode.

	var VERSION_STR string = "4.40"
	var DATE_STR string = "2007-01-16"

	var sign_on1 string = "go WavPack Encoder (c) 2011 Peter McQuillan\n"
	var sign_on2 string = "based on TINYPACK - Tiny Audio Compressor  Version " + VERSION_STR +
		" " + DATE_STR + " Copyright (c) 1998 - 2011 Conifer Software.  All Rights Reserved.\n"

	//////////////////////////////////////////////////////////////////////////////
	// The "main" function for the command-line WavPack compressor.             //
	//////////////////////////////////////////////////////////////////////////////
	var infilename string = ""
	var outfilename string = ""
	var out2filename string = ""
	config := new(wvencode.WavpackConfig)
	var error_count int = 0
	var result int
	var arg_idx int = 0
	var numArgs int = 0

	numArgs = len(os.Args)

	arg_idx = 1 // entry 0 is the program name

	// loop through command-line arguments
	for {
		if arg_idx >= numArgs {
			break
		}

		if os.Args[arg_idx][0] == '-' && len(os.Args[arg_idx]) > 1 {
			if os.Args[arg_idx][1] == 'c' || os.Args[arg_idx][1] == 'C' {
				if len(os.Args[arg_idx]) > 2 {
					if os.Args[arg_idx][2] == 'c' || os.Args[arg_idx][2] == 'C' {
						config.Flags = config.Flags | wvencode.CONFIG_CREATE_WVC
						config.Flags = config.Flags | wvencode.CONFIG_OPTIMIZE_WVC
					}
				} else {
					config.Flags = config.Flags | wvencode.CONFIG_CREATE_WVC
				}
			} else if os.Args[arg_idx][1] == 'f' || os.Args[arg_idx][1] == 'F' {
				config.Flags = config.Flags | wvencode.CONFIG_FAST_FLAG
			} else if os.Args[arg_idx][1] == 'h' || os.Args[arg_idx][1] == 'H' {
				if len(os.Args[arg_idx]) > 2 {
					if os.Args[arg_idx][2] == 'h' || os.Args[arg_idx][2] == 'H' {
						config.Flags = config.Flags | wvencode.CONFIG_VERY_HIGH_FLAG
					}
				} else {
					config.Flags = config.Flags | wvencode.CONFIG_HIGH_FLAG
				}
			} else if os.Args[arg_idx][1] == 'k' || os.Args[arg_idx][1] == 'K' {
				var passedInt int = 0

				if len(os.Args[arg_idx]) > 2 {
					var substring string = os.Args[arg_idx][2:len(os.Args[arg_idx])]
					pint, err := strconv.Atoi(substring)
					if err != nil {
						// Invalid string
					} else {
						passedInt = pint
					}
				} else {
					arg_idx++

					if arg_idx >= numArgs {
						break
					}

					pint, err := strconv.Atoi(os.Args[arg_idx])
					if err != nil {
						// Invalid string
					} else {
						passedInt = pint
					}
				}

				config.Block_samples = uint(passedInt)
			} else if os.Args[arg_idx][1] == 'b' || os.Args[arg_idx][1] == 'B' {

				config.Flags = config.Flags | wvencode.CONFIG_HYBRID_FLAG

				if len(os.Args[arg_idx]) > 2 { // handle the case where the string is passed in form -b0 (number beside b)
					var substring string = os.Args[arg_idx][2:len(os.Args[arg_idx])]
					pd, err := strconv.Atof64(substring)
					if err != nil {
						// Invalid string
						config.Bitrate = 0
					} else {
						config.Bitrate = int(math.Floor((pd * 256.0)))
					}
				} else {
					arg_idx++

					if arg_idx >= numArgs {
						break
					}

					pd, err := strconv.Atof64(os.Args[arg_idx])
					if err != nil {
						// Invalid string
						config.Bitrate = 0
					} else {
						config.Bitrate = int(math.Floor((pd * 256.0)))
					}
				}

				if (config.Bitrate < 512) || (config.Bitrate > 4096) {
					fmt.Printf("hybrid spec must be 2.0 to 16.0!\n")
					error_count++
				}
			} else if os.Args[arg_idx][1] == 'j' || os.Args[arg_idx][1] == 'J' {

				var passedInt int = 0

				if len(os.Args[arg_idx]) > 2 { // handle the case where the string is passed in form -j0 (number beside j)
					var substring string = os.Args[arg_idx][2:len(os.Args[arg_idx])]
					pint, err := strconv.Atoi(substring)
					if err != nil {
						// Invalid string
					} else {
						passedInt = pint
					}
				} else {
					arg_idx++

					if arg_idx >= numArgs {
						break
					}

					pint, err := strconv.Atoi(os.Args[arg_idx])
					if err != nil {
						// Invalid string
					} else {
						passedInt = pint
					}
				}

				if passedInt == 0 {
					config.Flags = config.Flags | wvencode.CONFIG_JOINT_OVERRIDE
					config.Flags = config.Flags & ^wvencode.CONFIG_JOINT_STEREO
				} else if passedInt == 1 {
					config.Flags = config.Flags | (wvencode.CONFIG_JOINT_OVERRIDE | wvencode.CONFIG_JOINT_STEREO)
				} else {
					fmt.Printf("-j0 or -j1 only!\n")
					error_count++
				}
			} else if os.Args[arg_idx][1] == 's' || os.Args[arg_idx][1] == 'S' {

				if len(os.Args[arg_idx]) > 2 { // handle the case where the string is passed in form -s0 (number beside s)
					var substring string = os.Args[arg_idx][2:len(os.Args[arg_idx])]
					pd, err := strconv.Atof64(substring)
					if err != nil {
						// Invalid string
						config.Shaping_weight = 0 // noise shaping off
					} else {
						config.Shaping_weight = int(math.Floor((pd * 1024.0)))
					}
				} else {
					arg_idx++

					if arg_idx >= numArgs {
						break
					}

					pd, err := strconv.Atof64(os.Args[arg_idx])
					if err != nil {
						// Invalid string
						config.Shaping_weight = 0
					} else {
						config.Shaping_weight = int(math.Floor((pd * 1024.0)))
					}
				}

				if config.Shaping_weight == 0 {
					config.Flags = config.Flags | wvencode.CONFIG_SHAPE_OVERRIDE
					config.Flags = config.Flags & ^wvencode.CONFIG_HYBRID_SHAPE
				} else if (config.Shaping_weight >= -1024) && (config.Shaping_weight <= 1024) {
					config.Flags = config.Flags | (wvencode.CONFIG_HYBRID_SHAPE | wvencode.CONFIG_SHAPE_OVERRIDE)
				} else {
					fmt.Printf("-s-1.00 to -s1.00 only!\n")
					error_count++
				}
			} else {
				fmt.Printf("illegal option: %s\n", os.Args[arg_idx])
				error_count++
			}
		} else if len(infilename) == 0 {
			infilename = os.Args[arg_idx]
		} else if len(outfilename) == 0 {
			outfilename = os.Args[arg_idx]
		} else if len(out2filename) == 0 {
			out2filename = os.Args[arg_idx]
		} else {
			fmt.Printf("extra unknown argument: %s\n", os.Args[arg_idx])
			error_count++
		}

		arg_idx++
		if arg_idx >= numArgs {
			break
		}
	}

	// check for various command-line argument problems
	if (^config.Flags & (wvencode.CONFIG_HIGH_FLAG | wvencode.CONFIG_FAST_FLAG)) == 0 {
		fmt.Printf("high and fast modes are mutually exclusive!\n")
		error_count++
	}

	if (config.Flags & wvencode.CONFIG_HYBRID_FLAG) != 0 {
		if ((config.Flags & wvencode.CONFIG_CREATE_WVC) != 0) && (len(out2filename) == 0) {
			fmt.Printf("need name for correction file!\n")
			error_count++
		}
	} else {
		if (config.Flags & (wvencode.CONFIG_SHAPE_OVERRIDE | wvencode.CONFIG_CREATE_WVC)) != 0 {
			fmt.Printf("-s and -c options are for hybrid mode (-b) only!\n")
			error_count++
		}
	}

	if (len(out2filename) != 0) && ((config.Flags & wvencode.CONFIG_CREATE_WVC) == 0) {
		fmt.Printf("third filename specified without -c option!\n")
		error_count++
	}

	if error_count == 0 {
		fmt.Printf(sign_on1)
		fmt.Printf(sign_on2)
	} else {
		os.Exit(1)
	}

	if (len(infilename) == 0) || (len(outfilename) == 0) ||
		((len(out2filename) == 0) && ((config.Flags & wvencode.CONFIG_CREATE_WVC) != 0)) {
		usage()
	}

	result = pack_file(infilename, outfilename, out2filename, config)

	if result > 0 {
		fmt.Printf("error occured!\n")
		error_count++
	}

}


// This function packs a single file "infilename" and stores the result at
// "outfilename". If "out2filename" is specified, then the "correction"
// file would go there. The files are opened and closed in this function
// and the "config" structure specifies the mode of compression.
func pack_file(infilename string, outfilename string, out2filename string, config *wvencode.WavpackConfig) int {
	var total_samples uint = 0
	var bcount int
	var loc_config *wvencode.WavpackConfig = config
	riff_chunk_header := make([]int, 12)
	chunk_header := make([]int, 8)

	var WaveHeader []int
	var whBlockAlign int = 1
	var whFormatTag int = 0
	var whSubFormat int = 0
	var whBitsPerSample int = 0
	var whValidBitsPerSample int = 0
	var whNumChannels uint = 0
	var whSampleRate uint = 0

	wpc := new(wvencode.WavpackContext)
	var result int

	din, err := os.Open(infilename)

	if err != nil {
		fmt.Printf("Cannot open input file %s\n", infilename)
		result = wvencode.HARD_ERROR
		return (result)
	}

	wv_file, err := os.OpenFile(outfilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		fmt.Printf("Error creating output file %s - error code is %s\n", outfilename, err.String())
		result = wvencode.HARD_ERROR
		return (result)
	}

	wpc.Outfile = wv_file

	bcount = 0

	// 12 is the size of the RIFF Chunk header
	bcount = DoReadFile(din, riff_chunk_header, 12)

	// ASCII values R = 82, I = 73, F = 70 (RIFF)
	// ASCII values W = 87, A = 65, V = 86, E = 69 (WAVE)
	if (bcount != 12) || (riff_chunk_header[0] != 82) || (riff_chunk_header[1] != 73) ||
		(riff_chunk_header[2] != 70) || (riff_chunk_header[3] != 70) ||
		(riff_chunk_header[8] != 87) || (riff_chunk_header[9] != 65) ||
		(riff_chunk_header[10] != 86) || (riff_chunk_header[11] != 69) {

		fmt.Printf("%s is not a valid .WAV file!\n", infilename)

		wv_file.Close()

		return wvencode.SOFT_ERROR
	}

	// loop through all elements of the RIFF wav header (until the data chuck)
	var chunkSize int = 0

	for {
		// ChunkHeader has a size of 8
		bcount = DoReadFile(din, chunk_header, 8)

		if bcount != 8 {
			fmt.Printf("%s is not a valid .WAV file!\n", infilename)

			wv_file.Close()

			return wvencode.SOFT_ERROR
		}

		chunkSize = (chunk_header[4] & 0xFF) + ((chunk_header[5] & 0xFF) << 8) +
			((chunk_header[6] & 0xFF) << 16) + ((chunk_header[7] & 0xFF) << 24)

		// if it's the format chunk, we want to get some info out of there and
		// make sure it's a .wav file we can handle
		// ASCII values f = 102, m = 109, t = 116, space = 32 ('fmt ')
		if (chunk_header[0] == 102) && (chunk_header[1] == 109) && (chunk_header[2] == 116) &&
			(chunk_header[3] == 32) {
			var supported int = wvencode.TRUE
			var format int
			var check int = 0

			if (chunkSize >= 16) && (chunkSize <= 40) {
				var ckSize int = chunkSize

				WaveHeader = make([]int, ckSize)

				bcount = DoReadFile(din, WaveHeader, ckSize)

				if bcount != ckSize {
					check = 1
				}
			} else {
				check = 1
			}

			if check == 1 {
				fmt.Printf("%s is not a valid .WAV file!\n", infilename)

				wv_file.Close()

				return wvencode.SOFT_ERROR
			}

			whFormatTag = (WaveHeader[0] & 0xFF) + ((WaveHeader[1] & 0xFF) << 8)

			if (whFormatTag == 0xe) && (chunkSize == 40) {
				whSubFormat = (WaveHeader[24] & 0xFF) + ((WaveHeader[25] & 0xFF) << 8)
				format = whSubFormat
			} else {
				format = whFormatTag
			}

			whBitsPerSample = (WaveHeader[14] & 0xFF) + ((WaveHeader[15] & 0xFF) << 8)

			if chunkSize == 40 {
				whValidBitsPerSample = (WaveHeader[18] & 0xFF) +
					((WaveHeader[19] & 0xFF) << 8)
				loc_config.Bits_per_sample = whValidBitsPerSample
			} else {
				loc_config.Bits_per_sample = whBitsPerSample
			}

			if format != 1 {
				supported = wvencode.FALSE
			}

			whBlockAlign = (WaveHeader[12] & 0xFF) + ((WaveHeader[13] & 0xFF) << 8)
			whNumChannels = uint((WaveHeader[2] & 0xFF) + ((WaveHeader[3] & 0xFF) << 8))

			if (whNumChannels == 0) || (whNumChannels > 2) ||
				(math.Floor(float64(whBlockAlign/int(whNumChannels))) < math.Floor(float64((loc_config.Bits_per_sample+7)/8))) ||
				(math.Floor(float64(whBlockAlign/int(whNumChannels))) > 3) ||
				((whBlockAlign % int(whNumChannels)) > 0) {
				supported = wvencode.FALSE
			}

			if (loc_config.Bits_per_sample < 1) || (loc_config.Bits_per_sample > 24) {
				supported = wvencode.FALSE
			}

			whSampleRate = uint((WaveHeader[4] & 0xFF) + ((WaveHeader[5] & 0xFF) << 8) +
				((WaveHeader[6] & 0xFF) << 16) + ((WaveHeader[7] & 0xFF) << 24))

			if supported != wvencode.TRUE {
				fmt.Printf("%s is an unsupported .WAV format!\n", infilename)

				wv_file.Close()

				return wvencode.SOFT_ERROR
			}
		} else if (chunk_header[0] == 100) && (chunk_header[1] == 97) &&
			(chunk_header[2] == 116) && (chunk_header[3] == 97) {
			// ASCII values d = 100, a = 97, t = 116
			// looking for string 'data'

			// on the data chunk, get size and exit loop
			total_samples = uint(math.Floor(float64(chunkSize / whBlockAlign)))

			break
		} else { // just skip over unknown chunks

			var bytes_to_skip int = ((chunkSize + 1) & ^1)
			var buff []int

			bcount = DoReadFile(din, buff, bytes_to_skip)

			if bcount != bytes_to_skip {
				fmt.Printf("error occurred in skipping bytes\n")

				wv_file.Close()

				//remove (outfilename);
				return wvencode.SOFT_ERROR
			}
		}
	}

	loc_config.Bytes_per_sample = int(math.Floor(float64(whBlockAlign / int(whNumChannels))))
	loc_config.Num_channels = whNumChannels
	loc_config.Sample_rate = whSampleRate

	wvencode.WavpackSetConfiguration(wpc, loc_config, total_samples)

	// if we are creating a "correction" file, open it now for writing
	if len(out2filename) > 0 {
		wvc_file, err := os.OpenFile(out2filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

		if err != nil {
			fmt.Printf("Cannot open output file %s\n", out2filename)
			result = wvencode.HARD_ERROR
			return (result)
		}

		wpc.Correction_outfile = wvc_file
	}

	// pack the audio portion of the file now
	result = pack_audio(wpc, din)

	din.Close() // we're now done with input file, so close

	// we're now done with any WavPack blocks, so flush any remaining data
	if (result == wvencode.NO_ERROR) && (wvencode.WavpackFlushSamples(wpc) == 0) {
		fmt.Printf("%s\n", wvencode.WavpackGetErrorMessage(wpc))
		result = wvencode.HARD_ERROR
	}

	// At this point we're done writing to the output files. However, in some
	// situations we might have to back up and re-write the initial blocks.
	// Currently the only case is if we're ignoring length.
	if (result == wvencode.NO_ERROR) &&
		(wvencode.WavpackGetNumSamples(wpc) != wvencode.WavpackGetSampleIndex(wpc)) {
		fmt.Printf("couldn't read all samples, file may be corrupt!!\n")
		result = wvencode.SOFT_ERROR
	}

	// at this point we're done with the files, so close 'em whether there
	// were any other errors or not

	errc := wv_file.Close()

	if errc != nil {
		if result == wvencode.NO_ERROR {
			result = wvencode.SOFT_ERROR
		}
	}

	// if there were any errors then return the error
	if result != wvencode.NO_ERROR {
		return result
	}

	return wvencode.NO_ERROR
}

// This function handles the actual audio data compression. It assumes that the
// input file is positioned at the beginning of the audio data and that the
// WavPack configuration has been set. This is where the conversion from RIFF
// little-endian standard the executing processor's format is done.
func pack_audio(wpc *wvencode.WavpackContext, din *os.File) int {
	var samples_remaining int
	var bytes_per_sample int

	wvencode.WavpackPackInit(wpc)

	bytes_per_sample = wvencode.WavpackGetBytesPerSample(wpc) * wvencode.WavpackGetNumChannels(wpc)

	samples_remaining = wvencode.WavpackGetNumSamples(wpc)

	var input_buffer []int
	var sample_buffer []int

	var temp int = 0

	for {
		var sample_count uint
		var bytes_read int = 0
		var bytes_to_read int

		temp = temp + 1

		if samples_remaining > wvencode.INPUT_SAMPLES {
			bytes_to_read = wvencode.INPUT_SAMPLES * bytes_per_sample
		} else {
			bytes_to_read = (samples_remaining * bytes_per_sample)
		}

		samples_remaining -= int(math.Floor(float64(bytes_to_read / bytes_per_sample)))

		input_buffer = make([]int, bytes_to_read)
		bytes_read = DoReadFile(din, input_buffer, bytes_to_read)

		sample_count = uint(math.Floor(float64(bytes_read / bytes_per_sample)))

		if sample_count == 0 {
			break
		}

		if sample_count > 0 {
			var cnt int = (int(sample_count) * wvencode.WavpackGetNumChannels(wpc))

			sptr := input_buffer[0:len(input_buffer)]

			var loopBps int = 0

			loopBps = wvencode.WavpackGetBytesPerSample(wpc)

			if loopBps == 1 {

				var internalCount int = 0

				sample_buffer = make([]int, cnt)

				sample_buffer[cnt-1] = 0 // initialize array
				for cnt > 0 {
					sample_buffer[internalCount] = (sptr[internalCount] & 0xff) - 128
					internalCount++
					cnt--
				}
			} else if loopBps == 2 {

				var dcounter int = 0
				var scounter int = 0

				sample_buffer = make([]int, cnt)
				sample_buffer[cnt-1] = 0 // initialize array
				for cnt > 0 {
					sample_buffer[dcounter] = (sptr[scounter] & 0xff) | (sptr[scounter+1] << 8)

					scounter = scounter + 2
					dcounter++
					cnt--
				}
			} else if loopBps == 3 {

				var dcounter int = 0
				var scounter int = 0

				sample_buffer = make([]int, cnt)
				sample_buffer[cnt-1] = 0 // initialize array
				for cnt > 0 {
					sample_buffer[dcounter] = (sptr[scounter] & 0xff) |
						((sptr[scounter+1] & 0xff) << 8) | (sptr[scounter+2] << 16)
					scounter = scounter + 3
					dcounter++
					cnt--
				}
			}
		}

		wpc.Byte_idx = 0 // new WAV buffer data so reset the buffer index to zero

		if wvencode.WavpackPackSamples(wpc, sample_buffer, sample_count) == 0 {
			fmt.Printf("%s\n", wvencode.WavpackGetErrorMessage(wpc))

			return wvencode.HARD_ERROR
		}

	}

	if wvencode.WavpackFlushSamples(wpc) == 0 {

		fmt.Printf("%s\n", wvencode.WavpackGetErrorMessage(wpc))

		return wvencode.HARD_ERROR
	}

	return wvencode.NO_ERROR
}

//////////////////////////// File I/O Wrapper ////////////////////////////////
func DoReadFile(hFile *os.File, lpBuffer []int, nNumberOfBytesToRead int) int {
	tempBufferAsBytes := make([]byte, nNumberOfBytesToRead)

	var lpNumberOfBytesRead int = 0
	var tempI int = 0

	for nNumberOfBytesToRead > 0 {
		bcount, inErr := hFile.Read(tempBufferAsBytes)

		if inErr != nil {
			fmt.Printf("Error encountered\n")
		}

		if bcount > 0 {
			lpBuffer[bcount-1] = 0
			for i := 0; i < bcount; i++ {
				tempI = int(tempBufferAsBytes[i])
				// the following is a very inelegant way to convert unsigned to signed bytes
				// must be a better way in go!
				if tempI > 127 {
					tempI = tempI - 256
				}
				lpBuffer[i] = tempI
				//				if(i % 1000 == 0) {
				//					fmt.Printf(".");
				//				}
			}

			lpNumberOfBytesRead += bcount
			nNumberOfBytesToRead -= bcount
		} else {
			break
		}
	}

	return lpNumberOfBytesRead
}
