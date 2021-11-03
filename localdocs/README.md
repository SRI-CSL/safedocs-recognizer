# Location for local corpus files

sed -i -e 's/\/localdisk01\/corpora\///g' evalThree_10K.txt

... make relative to symlink in this directory

cp universeA-filelist.txt sd_index
gzip sd_index