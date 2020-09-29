cmds = chihiro
vfile = pkg/version/version.go
vdata = `git describe --tags`-`date -u +%Y%m%d%H%M%S`

.PHONY: $(cmds) all clean version tidy package

all: $(cmds)

clean:
	rm -f $(cmds) $(addsuffix .exe, $(cmds))

$(cmds): version tidy
	go build ./cmd/$@

version:
	rm -f $(vfile)
	@echo "package version" > $(vfile)
	@echo "const (" >> $(vfile)
	@echo "  Version = \"$(vdata)\"" >> $(vfile)
	@echo ")" >> $(vfile)

tidy:
	go mod tidy && go mod vendor && go fmt ./pkg/* ./cmd/*
