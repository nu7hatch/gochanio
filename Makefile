include $(GOROOT)/src/Make.inc

TARG=chanio
GOFILES= \
	chanio.go \
	doc.go

include $(GOROOT)/src/Make.pkg
