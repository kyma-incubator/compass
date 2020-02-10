res=0

echo "Run Provisioner tests"
./provisioner.test -test.v -test.timeout 120m
res=$((res+$?))

exit ${res}