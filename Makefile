all:
	mkdir -p bin
	cd stts && CGO_ENABLED=0 go build -o ../bin/stts 
	cd mailcount && CGO_ENABLED=0 go build -o ../bin/mailcount
	cd airvpn && CGO_ENABLED=0 go build -o ../bin/airvpn
	cd string_normalize && CGO_ENABLED=0 go build -o ../bin/string_normalize

