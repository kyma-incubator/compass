res=0

echo "Run Provisioner tests"
./apitests.test -test.v
res=$((res+$?))

exit ${res}