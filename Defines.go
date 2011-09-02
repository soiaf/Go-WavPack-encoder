package wvencode

/*
** Defines.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

const BIT_BUFFER_SIZE = 65536 // This should be carefully chosen for the
// application and platform. Larger buffers are
// somewhat more efficient, but the code will
// allow smaller buffers and simply terminate
// blocks early. If the hybrid lossless mode
// (2 file) is not needed then the wvc_buffer
// can be made very small.
// or-values for "flags"
const INPUT_SAMPLES int = 65536
const BYTES_STORED uint = 3                  // 1-4 bytes/sample
const CONFIG_AUTO_SHAPING uint = 0x4000      // automatic noise shaping
const CONFIG_BITRATE_KBPS uint = 0x2000      // bitrate is kbps, not bits / sample
const CONFIG_BYTES_STORED uint = 3           // 1-4 bytes/sample
const CONFIG_CALC_NOISE uint = 0x800000      // calc noise in hybrid mode
const CONFIG_CREATE_EXE uint = 0x40000       // create executable
const CONFIG_CREATE_WVC uint = 0x80000       // create correction file
const CONFIG_CROSS_DECORR uint = 0x20        // no-delay cross decorrelation
const CONFIG_EXTRA_MODE uint = 0x2000000     // extra processing mode
const CONFIG_FAST_FLAG uint = 0x200          // fast mode
const CONFIG_FLOAT_DATA uint = 0x80          // ieee 32-bit floating point data
const CONFIG_HIGH_FLAG uint = 0x800          // high quality mode
const CONFIG_HYBRID_FLAG uint = 8            // hybrid mode
const CONFIG_HYBRID_SHAPE uint = 0x40        // noise shape (hybrid mode only)
const CONFIG_JOINT_OVERRIDE uint = 0x10000   // joint-stereo mode specified
const CONFIG_JOINT_STEREO uint = 0x10        // joint stereo
const CONFIG_LOSSY_MODE uint = 0x1000000     // obsolete (for information)
const CONFIG_MD5_CHECKSUM uint = 0x8000000   // compute & store MD5 signature
const CONFIG_MONO_FLAG uint = 4              // not stereo
const CONFIG_OPTIMIZE_MONO uint = 0x80000000 // optimize for mono streams posing as stereo
const CONFIG_OPTIMIZE_WVC uint = 0x100000    // maximize bybrid compression
const CONFIG_SHAPE_OVERRIDE uint = 0x8000    // shaping mode specified
const CONFIG_SKIP_WVX uint = 0x4000000       // no wvx stream w/ floats & big ints
const CONFIG_VERY_HIGH_FLAG uint = 0x1000    // very high
const CROSS_DECORR uint = 0x20               // no-delay cross decorrelation
const CUR_STREAM_VERS int = 0x405            // stream version we are writing now

// encountered
const FALSE int = 0
const FALSE_STEREO uint = 0x40000000 // block is stereo, but data is mono
const FINAL_BLOCK uint = 0x1000      // final block of multichannel segment
const FLOAT_DATA int = 0x80          // ieee 32-bit floating point data
const FLOAT_EXCEPTIONS int = 0x20    // contains exceptions (inf, nan, etc.)
const FLOAT_NEG_ZEROS int = 0x10     // contains negative zeros
const FLOAT_SHIFT_ONES int = 1       // bits left-shifted into float = '1'
const FLOAT_SHIFT_SAME int = 2       // bits left-shifted into float are the same
const FLOAT_SHIFT_SENT int = 4       // bits shifted into float are sent literally
const FLOAT_ZEROS_SENT int = 8       // "zeros" are not all real zeros
const HARD_ERROR int = 2
const HYBRID_BALANCE uint = 0x400 // balance noise (hybrid stereo mode only)
const HYBRID_BITRATE uint = 0x200 // bitrate noise (hybrid mode only)
const HYBRID_FLAG uint = 8        // hybrid mode
const HYBRID_SHAPE uint = 0x40    // noise shape (hybrid mode only)
const ID_CHANNEL_INFO uint = 0xd
const ID_CONFIG_BLOCK int = 0x25
const ID_CUESHEET uint = 0x24
const ID_DECORR_SAMPLES int = 0x4
const ID_DECORR_TERMS int = 0x2
const ID_DECORR_WEIGHTS int = 0x3
const ID_DUMMY uint = 0x0
const ID_ENCODER_INFO uint = 0x1
const ID_ENTROPY_VARS int = 0x5
const ID_FLOAT_INFO uint = 0x8
const ID_HYBRID_PROFILE int = 0x6
const ID_INT32_INFO uint = 0x9
const ID_LARGE int = 0x80
const ID_MD5_CHECKSUM uint = 0x26
const ID_ODD_SIZE int = 0x40
const ID_OPTIONAL_DATA uint = 0x20
const ID_REPLAY_GAIN uint = 0x23
const ID_RIFF_HEADER uint = 0x21
const ID_RIFF_TRAILER uint = 0x22
const ID_SAMPLE_RATE int = 0x27
const ID_SHAPING_WEIGHTS int = 0x7
const ID_WVC_BITSTREAM int = 0xb
const ID_WVX_BITSTREAM uint = 0xc
const ID_WV_BITSTREAM int = 0xa
const IGNORED_FLAGS int = 0x18000000 // reserved, but ignore if encountered
const INITIAL_BLOCK uint = 0x800     // initial block of multichannel segment
const INT32_DATA int = 0x100         // special extended int handling
const JOINT_STEREO uint = 0x10       // joint stereo
const MAG_LSB uint = 18
const MAX_NTERMS int = 16
const MAX_STREAM_VERS int = 0x410 // highest stream version we'll decode
const MAX_TERM = 8

const MIN_STREAM_VERS int = 0x402 // lowest stream version we'll decode
const MODE_FAST int = 0x40
const MODE_FLOAT int = 0x8
const MODE_HIGH int = 0x20
const MODE_HYBRID int = 0x4
const MODE_LOSSLESS int = 0x2
const MODE_VALID_TAG int = 0x10
const MODE_WVC int = 0x1
const MONO_FLAG uint = 4            // not stereo
const NEW_SHAPING uint = 0x20000000 // use IIR filter for negative shaping
const NO_ERROR int = 0

// Change the following value to an even number to reflect the maximum number of samples to be processed
// per call to WavPackUtils.WavpackUnpackSamples
const SAMPLE_BUFFER_SIZE int = 256
const SHIFT_LSB uint = 13
const SOFT_ERROR int = 1
const SRATE_LSB uint = 23
const TRUE int = 1
const UNKNOWN_FLAGS uint = 0x80000000 // also reserved, but refuse decode if
const WAVPACK_HEADER_SIZE int = 32
const SRATE_MASK uint = (0xf << SRATE_LSB)
const SHIFT_MASK uint = (0x1f << SHIFT_LSB)
const MAG_MASK uint = (0x1f << MAG_LSB)
