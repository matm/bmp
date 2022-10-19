.PHONY: bmp dist linux darwin windows buildall cleardist clean

include version.mk

BIN=bmp
WINDOWS_BIN=${BIN}.exe
DISTDIR=dist
BINDIR=${DISTDIR}/${BIN}
BUILDDIR=build
#
BUILD_VERSION=${BIN}-${VERSION}
BUILD_DARWIN_AMD64=${BUILD_VERSION}-darwin-amd64
BUILD_FREEBSD_AMD64=${BUILD_VERSION}-freebsd-amd64
BUILD_OPENBSD_AMD64=${BUILD_VERSION}-openbsd-amd64
BUILD_LINUX_AMD64=${BUILD_VERSION}-linux-amd64
BUILD_WINDOWS_AMD64=${BUILD_VERSION}-windows-amd64
#
GOBUILD64=GOARCH=amd64 go build
MAIN_CMD=github.com/matm/${BIN}/cmd/${BIN}

all: build

build:
	@go build -ldflags "all=$(GO_LDFLAGS)" ${MAIN_CMD}

dist: cleardist buildall zip sourcearchive checksum

test:
	@go test -v ./...

checksum:
	@for f in ${DISTDIR}/*; do \
		sha256sum $$f > $$f.sha256; \
		sed -i 's,${DISTDIR}/,,' $$f.sha256; \
	done

zip: linux freebsd windows
	@rm -rf ${BINDIR}

linux:
	@cp ${BUILDDIR}/${BUILD_VERSION}-linux* ${BINDIR}/${BIN} && \
		(cd ${DISTDIR} && zip -r ${BUILD_LINUX_AMD64}.zip ${BIN})

darwin:
	@cp ${BUILDDIR}/${BUILD_VERSION}-darwin* ${BINDIR}/${BIN} && \
		(cd ${DISTDIR} && zip -r ${BUILD_DARWIN_AMD64}.zip ${BIN})

windows:
	@cp ${BUILDDIR}/${BUILD_VERSION}-windows* ${BINDIR}/${WINDOWS_BIN} && \
		(cd ${DISTDIR} && rm ${BIN}/${BIN} && zip -r ${BUILD_WINDOWS_AMD64}.zip ${BIN})

freebsd:
	@cp ${BUILDDIR}/${BUILD_VERSION}-freebsd* ${BINDIR}/${BIN} && \
		(cd ${DISTDIR} && zip -r ${BUILD_FREEBSD_AMD64}.zip ${BIN})

openbsd:
	@cp ${BUILDDIR}/${BUILD_VERSION}-openbsd* ${BINDIR}/${BIN} && \
		(cd ${DISTDIR} && zip -r ${BUILD_OPENBSD_AMD64}.zip ${BIN})

buildall:
#@GOOS=darwin ${GOBUILD64} -v -o ${BUILDDIR}/${BUILD_DARWIN_AMD64} ${MAIN_CMD}
#@GOOS=openbsd ${GOBUILD64} -v -o ${BUILDDIR}/${BUILD_OPENBSD_AMD64} ${MAIN_CMD}
	@GOOS=freebsd ${GOBUILD64} -v -o ${BUILDDIR}/${BUILD_FREEBSD_AMD64} ${MAIN_CMD}
	@GOOS=linux ${GOBUILD64} -v -o ${BUILDDIR}/${BUILD_LINUX_AMD64} ${MAIN_CMD}
	@GOOS=windows ${GOBUILD64} -v -o ${BUILDDIR}/${BUILD_WINDOWS_AMD64} ${MAIN_CMD}

sourcearchive:
	@git archive --format=zip -o ${DISTDIR}/${BUILD_VERSION}.zip ${VERSION}
	@echo ${DISTDIR}/${BUILD_VERSION}.zip
	@git archive -o ${DISTDIR}/${BUILD_VERSION}.tar ${VERSION}
	@gzip ${DISTDIR}/${BUILD_VERSION}.tar
	@echo ${DISTDIR}/${BUILD_VERSION}.tar.gz

cleardist:
	@rm -rf ${DISTDIR} && mkdir -p ${BINDIR} && mkdir -p ${BUILDDIR}

clean:
	@rm -rf ${BIN} ${BUILDDIR} ${DISTDIR}
