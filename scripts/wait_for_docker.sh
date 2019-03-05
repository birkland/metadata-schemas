#!/bin/sh

MAX_ATTEMPTS=30

if [ -z ${PASS_FEDORA_USER+x} ]; then 
    PASS_FEDORA_USER="fedoraAdmin"
fi

if [ -z ${PASS_FEDORA_PASSWORD+x} ]; then 
    PASS_FEDORA_PASSWORD="moo"
fi

if [ -z ${PASS_EXTERNAL_FEDORA_BASEURL+x} ]; then 
    PASS_EXTERNAL_FEDORA_BASEURL="http://localhost:8080/fcrepo/rest"
fi

if [ -z ${SCHEMA_SERVICE_URL+x} ]; then 
    SCHEMA_SERVICE_URL="http://localhost:8086"
fi

function wait_until_up {
    CMD="curl -I -u ${PASS_FEDORA_USER}:${PASS_FEDORA_PASSWORD} --write-out %{http_code} --silent -o /dev/stderr ${1}"
    echo "Waiting for response from via ${CMD}"

    RESULT=0
    max=${MAX_ATTEMPTS}
    i=1
    
    until [ ${RESULT} -eq 200 ]
    do
        sleep 5
        
    RESULT=$(${CMD})

        if [ $i -eq $max ]
        then
            echo "Reached max attempts"
            exit 1
        fi

        i=$((i+1))
        echo "Trying again, result was ${RESULT}"
    done
    
    echo "$1 is up."
}

wait_until_up ${PASS_EXTERNAL_FEDORA_BASEURL}
wait_until_up ${SCHEMA_SERVICE_URL}