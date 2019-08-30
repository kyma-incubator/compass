res=0

echo "Run Connector tests"
./apitests.test -test.v
res=$((res+$?))

exit ${res}