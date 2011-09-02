package wvencode

/*
** DecorrPass.go
**
** Copyright (c) 2011 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type DecorrPass struct {
	term      int
	delta     int
	weight_A  int
	weight_B  int
	samples_A [MAX_TERM]int
	samples_B [MAX_TERM]int
	aweight_A int
	aweight_B int
}
