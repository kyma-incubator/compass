res=0

echo "Run Kyma environment broker tests"
./test.test -test.v
res=$((res+$?))

exit ${res}