all: build pack

build:
	go build

pack:
	tar czvf moduleab_agent.tar.gz moduleab_agent conf.ini logs --excldue=logs/*

clean:
	rm moduleab_agent || echo
	rm logs/* || echo
