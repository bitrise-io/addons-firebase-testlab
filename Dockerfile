FROM quay.io/bitriseio/bitrise-base

# envs
ENV PROJ_NAME=addons-firebase-testlab
ENV BITRISE_SOURCE_DIR="/bitrise/go/src/github.com/bitrise-io/$PROJ_NAME"

# Get go tools
RUN go get github.com/codegangsta/gin \
    && go get github.com/kisielk/errcheck \
    && go get golang.org/x/lint/golint \
    && go get github.com/stripe/safesql \
    && go get github.com/gobuffalo/buffalo/buffalo

RUN apt-get update -qq && apt-get upgrade -y && \
    apt-get install -y postgresql-client

WORKDIR $BITRISE_SOURCE_DIR

CMD $PROJ_NAME
