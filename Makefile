versionstr := rel-$(shell cat version.txt)

build:
	go build

test:
	TZ=Z go test

tag-release:
	git tag -a -s -m $(versionstr) $(versionstr)
