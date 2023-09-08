build:
	go build -o pix .

install: build
	chmod +x pix
	mv pix ~/bin
