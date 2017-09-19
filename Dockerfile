FROM golang:1.9.0

ENV LC_ALL=C.UTF-8 LANG=C.UTF-8 PIPENV_VENV_IN_PROJECT=1

RUN apt-get update && \
    apt-get -y install python3-pip && \
    pip3 install pipenv

RUN useradd -ms /bin/bash golang

WORKDIR /home/golang/goot

COPY tests/Pipfile tests/Pipfile.lock ./tests/
RUN cd tests && pipenv install

COPY . .
RUN go build

WORKDIR /home/golang/goot/tests

RUN chown -R golang /home/golang
USER golang

CMD ["pipenv", "run", "pytest", ".", "--", "--docker"]
