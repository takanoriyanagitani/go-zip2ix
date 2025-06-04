#!/bin/sh

izfile="./sample.d/input.zip"
oasn1="./sample.d/output.asn1.der.dat"

geninput(){
	echo creating input zip file...

	mkdir -p ./sample.d

	printf hw0 > ./sample.d/f0.txt
	printf hw1 > ./sample.d/f1.txt
	printf hwii > ./sample.d/f2.txt
	printf hwiii > ./sample.d/f3.txt
	printf hwiv > ./sample.d/f4.txt

	find \
		./sample.d \
		-type f \
		-name '*.txt' |
		sort |
		zip \
			-0 \
			-@ \
			-T \
			-v \
			-o \
			"${izfile}"
}

test -f "${izfile}" || geninput


echo
echo creating an index file from the zip file...
wazero \
	run \
	-env ENV_INPUT_ZIP_FILENAME=/guest-i.d/input.zip \
	-mount "${PWD}/sample.d:/guest-i.d:ro" \
	./zip2ix2asn1.wasm |
	dd \
		if=/dev/stdin \
		of="${oasn1}" \
		bs=1048576 \
		status=none


echo
echo converting DER to JER using asn1tools and jq...
cat "${oasn1}" |
	xxd -ps |
	tr -d '\n' |
	python3 \
		-m asn1tools \
		convert \
		-i der \
		-o jer \
		./least_zip_ix_info.asn \
		SequenceOfIndexInfo \
		- |
	jq -c '.[]'


iseek=302
count=5

iseek=149
count=3

echo
echo printing the 2nd contents using the offset/size info...
dd \
	if="${izfile}" \
	of=/dev/stdout \
	bs=1 \
	skip=$iseek \
	count=$count \
	status=none |
	xxd
