res=0

echo "Run Provisioner tests"
./provisioner.test -test.v
res=$((res+$?))

exit ${res}