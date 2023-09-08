#!/bin/bash

src_dir="$1"
dst_dir="$2"

exiftool -overwrite_original -tagsfromfile "$src_dir/%f.%e" "$dst_dir"
