#!/bin/sh

INPUT_COUNT=3000
#INPUT_COUNT=60
#COMMAND="./fetchstdin"
COMMAND="./kamehame -conc 1000 -tps 300"

url_list()
{
    cat <<EOF | awk 'sub(/#.*/,"")>=0&&NF>0'
GET	http://127.0.0.1:8000/	NULL
GET	http://127.0.0.1:8000/aaa	NULL
GET	http://127.0.0.1:8000/bbb	NULL
POST	http://127.0.0.1:8000/bbb	./tmpl/request1.tmpl
#POST	http://127.0.0.1:8000/ccc	./tmpl/request2.tmpl
#GET	https://www.google.co.jp/	NULL
EOF
}

url_input()
{
    local loop=$(( $INPUT_COUNT/$(url_list|wc -l) ))
	local URL_LIST=$(url_list)
    for i in $(seq $loop); do
        #url_list
		echo "$URL_LIST"
    done
    url_list | head -$(( $INPUT_COUNT%$(url_list|wc -l) ))
}

#url_list | ./fetchstdin
#url_input|wc -l
#url_input | ./fetchstdin 
#url_input | ./fetchstdin | awk '{count[$1]++;} NR%100==0{for (key in count){ print key,count[key];}}'
#url_input | ./concfetch | awk '{count[$1]++;} NR%100==0{for (key in count){ print key,count[key];}}'

#url_input | eval $COMMAND
#exit
url_input | eval $COMMAND | awk '
    {
        count[$1]++;
    }
    NR%100==0 {
        for (key in count){
            #print strftime("%y/%m/%d %H:%M:%S"),key,count[key];
            print strftime("%c"),key,count[key];
			#fflush()
        }
    }'
