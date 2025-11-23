TEMPL := $(shell command -v templ 2> /dev/null)
GOLANG := $(shell command -v go 2> /dev/null)
outputfile := bin/todo 

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
	outputfile := "bin/todo.exe"
else
	outputfile := "bin/todo" 
endif

common_test:
ifndef GOLANG
	$(error "need to have go installed. please install: https://go.dev/learn/ ")
endif
ifndef TEMPL
	$(error "need to install templ. please install: https://templ.guide/quick-start/installation ")
endif
	go mod tidy

build: common_test
	templ generate
	go build -o $(outputfile)

run: common_test	 
	templ generate
	go run .

dev: common_test
	templ generate --watch --proxy="http://localhost:3010" --cmd="go run ."

clean:
	rm ./pages/*.go
	rm ./template/*.go
	rm ./bin/*.*
	