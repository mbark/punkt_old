FROM golang:1.9.0

ENV LC_ALL=C.UTF-8 \
    LANG=C.UTF-8 \
    PIPENV_VENV_IN_PROJECT=1 \
    PROJECT=github.com/mbark/goot

RUN apt-get update && \
    apt-get -y install python3-pip && \
    pip3 install pipenv

RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/$PROJECT
WORKDIR /go/src/$PROJECT

COPY Gopkg.lock Gopkg.toml ./
RUN dep ensure -vendor-only

COPY tests/Pipfile tests/Pipfile.lock ./tests/
RUN cd tests && pipenv install

COPY . .
RUN go build

WORKDIR ./tests

CMD ["pipenv", "run", "pytest", ".", "--", "--docker"]
