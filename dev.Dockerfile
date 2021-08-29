FROM golang:1.17.0-stretch

RUN apt update && apt upgrade -y && \
    apt install -y git \
    make openssh-client

WORKDIR /app 

# Install Air
RUN curl -fLo install.sh https://raw.githubusercontent.com/cosmtrek/air/master/install.sh \
    && chmod +x install.sh && sh install.sh \
    && cp ./bin/air /bin/air \
    # Install go-migrate
    && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

CMD air
