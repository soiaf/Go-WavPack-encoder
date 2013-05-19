package wvencode

/*
** WordsData.go
**
** Copyright (c) 2013 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type WordsData struct {
	bitrate_delta []int  // was uint32_t  in C
	bitrate_acc   []uint // was uint32_t  in C
	pend_data     uint   // was uint32_t  in C
	holding_one   uint   // was uint32_t  in C
	zeros_acc     uint   // was uint32_t  in C
	median        [3][2]int
	slow_level    []int // was uint32_t  in C
	error_limit   []int // was uint32_t  in C
	holding_zero  int
	pend_count    uint
}
