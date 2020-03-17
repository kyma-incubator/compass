res=0

echo "Run Provisioner tests"
./provisioner.test -test.v -test.timeout 2h
res=$((res+$?))

exit ${res}