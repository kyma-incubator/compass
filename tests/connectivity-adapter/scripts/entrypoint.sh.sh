res=0

echo "Run Connectivity Adapter tests"
./apitests.test -test.v
res=$((res+$?))

exit ${res}