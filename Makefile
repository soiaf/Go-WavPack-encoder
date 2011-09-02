include $(GOROOT)/src/Make.inc

TARG=wvencode
GOFILES=\
	Defines.go\
	WavpackConfig.go\
	WavpackContext.go\
	WavPackUtils.go\
	WavpackStream.go\
	DecorrPass.go\
	WavpackHeader.go\
	Bitstream.go\
	DeltaData.go\
	WordsData.go\
	PackUtils.go\
	WavpackMetadata.go\
	BitsUtils.go\
	WordsUtils.go\

include $(GOROOT)/src/Make.pkg

