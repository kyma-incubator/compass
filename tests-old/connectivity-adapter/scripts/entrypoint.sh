res=0

echo "Run Connectivity Adapter tests"
./api.test -test.v
res=$((res+$?))

exit ${res}