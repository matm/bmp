VERSION = $(shell git describe --tags)
GITREV = $(shell git rev-parse --verify --short HEAD)
GITBRANCH = $(shell git rev-parse --abbrev-ref HEAD)
DATE = $(shell LANG=US date +"%a, %d %b %Y %X %z")

GO_LDFLAGS += -X 'github.com/matm/bmp/pkg/config.Version=$(VERSION)'
GO_LDFLAGS += -X 'github.com/matm/bmp/pkg/config.GitRev=$(GITREV)'
GO_LDFLAGS += -X 'github.com/matm/bmp/pkg/config.GitBranch=$(GITBRANCH)'
GO_LDFLAGS += -X 'github.com/matm/bmp/pkg/config.BuildDate=$(DATE)'
