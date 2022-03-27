
PREFIX		=	${HOME}/.local
SHAREDIR	=	${PREFIX}/share/gamenode
CONFDIR		=	${PREFIX}/etc
GOBIN		=	${PREFIX}/bin

GO111MODULE	=	auto

all: proto build

build:
	go build ./cmd/gamenode

install: proto build 
	go env -w GOBIN=${GOBIN}
	go install ./cmd/gamenode
	mkdir -p ${SHAREDIR}
	cp -f etc/gamenode.conf ${CONFDIR}
	cp -f gamenode.service ${HOME}/.config/systemd/user
	systemctl --user daemon-reload

proto: pkg/gamenodepb/gamenodepb.pb.go

pkg/gamenodepb/gamenodepb.pb.go: pkg/gamenodepb/gamenodepb.proto
	protoc  --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			pkg/gamenodepb/gamenodepb.proto
