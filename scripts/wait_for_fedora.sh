#!/bin/sh

FCREPO_MAX_ATTEMPTS=30

if [ -z ${PASS_FEDORA_USER+x} ]; then 
    PASS_FEDORA_USER="fedoraAdmin"
fi

if [ -z ${PASS_FEDORA_PASSWORD+x} ]; then 
    PASS_FEDORA_PASSWORD="moo"
fi

if [ -z ${PASS_EXTERNAL_FEDORA_BASEURL+x} ]; then 
    echo "setting up Fedora baseurl"
    PASS_EXTERNAL_FEDORA_BASEURL="http://localhost:8080/fcrepo/rest"
else 
    echo "NOT setting up baseurl"
fi

CMD="curl -I -u ${PASS_FEDORA_USER}:${PASS_FEDORA_PASSWORD} --write-out %{http_code} --silent -o /dev/stderr ${PASS_EXTERNAL_FEDORA_BASEURL}"
echo "Waiting for response from Fedora via ${CMD}"

RESULT=0
max=${FCREPO_MAX_ATTEMPTS}
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
    
echo "Fedora is up."