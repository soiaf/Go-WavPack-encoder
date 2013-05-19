package wvencode

/*
** Bitstream.go
**
** Copyright (c) 2013 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type Bitstream struct {
	end         int // was uchar in c
	sr          uint
	error       int
	bc          uint
	buf_index   int
	start_index int
	active      int // if 0 then this bitstream is not being used
}
