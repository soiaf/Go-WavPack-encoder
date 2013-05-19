package wvencode

/*
** WavpackMetadata.go
**
** Copyright (c) 2013 Peter McQuillan
**
** All Rights Reserved.
**
** Distributed under the BSD Software License (see license.txt)
**
 */

type WavpackMetadata struct {
	byte_length int
	temp_data   [64]byte
	data        []byte
	id          int
}
