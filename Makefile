all:
	docker build -t faryon93/nsqflux:latest .

push:
	docker push faryon93/nsqflux:latest
