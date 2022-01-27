
PREFIX		=	${HOME}/.local
SHAREDIR	=	${PREFIX}/share/ednode
CONFDIR		=	${PREFIX}/etc
GOBIN		=	${PREFIX}/bin

GO111MODULE	=	auto

all:
	go build ./cmd/ednode

install:
	go env -w GOBIN=${GOBIN}
	go install ./cmd/ednode
	mkdir -p ${SHAREDIR}
	cp -f etc/ednode.conf ${CONFDIR}

