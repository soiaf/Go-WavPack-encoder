package wvencode

/*
** WavpackConfig.go
**
** Copyright (c) 2013 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type WavpackConfig struct {
	Bitrate          int
	Shaping_weight   int
	Bits_per_sample  int
	Bytes_per_sample int
	Num_channels     uint
	Block_samples    uint
	Flags            uint
	Sample_rate      uint
}
