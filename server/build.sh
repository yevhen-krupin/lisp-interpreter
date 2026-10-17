go -C ../core test \
&& go -C ../core build  \
&& go build \
&& git describe --tags --abbrev=0 > ./static/version.txt 