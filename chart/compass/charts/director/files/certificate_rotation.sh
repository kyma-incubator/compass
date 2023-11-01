RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

set -e

IS_CERT_ISSUER_LOCALITY_VALID=false

trap "echo '# KILLING PILOT-AGENT #'; curl -X POST http://127.0.0.1:15020/quitquitquit; sleep 3;" RETURN EXIT INT TERM

function getClientCert() {
  echo -e "${YELLOW}Getting the certificate chain... ${NC}"
  CERT_CHAIN_PKCS7_RESP=$(curl -s -m 30 -X POST \
    "$CERT_SVC_CSR_ENDPOINT$CERT_SVC_API_PATH" \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    -d "{
          \"certificate-signing-request\": {
              \"value\": $JSON_CSR,
              \"type\": \"pkcs10-pem\",
              \"validity\": {
                \"value\": $CERT_VALIDITY,
                \"type\": \"DAYS\"
            }
          }
        }")

  echo -e "${YELLOW}JSON-decoding client certificate chain... ${NC}"
  CERT_CHAIN_PKCS7=$(jq -r '.["certificate-response"]["value"]' <<< "$CERT_CHAIN_PKCS7_RESP")

  if [ "$CERT_CHAIN_PKCS7" == "null" ]
  then
    echo -e "${RED}Could not get certificate response. Reason: ${NC}"
    echo "$CERT_CHAIN_PKCS7_RESP"
    exit 1
  fi

  echo -e "${YELLOW}Extracting client certificate... ${NC}"
  openssl pkcs7 -print_certs -out /tmp/client-certificate_pkcs7.pem <<< "$CERT_CHAIN_PKCS7"
  openssl x509 -in /tmp/client-certificate_pkcs7.pem -out /tmp/client-certificate.pem
}

function confirmValidIssuerLocalityOrRetry() {
  for (( i = 0; i < "$CLIENT_CERT_RETRY_ATTEMPTS"; i++ )); do
    echo -e "${YELLOW}Checking issuer locality... ${NC}"
    ISSUER_LOCALITY=$(openssl x509 -in /tmp/client-certificate.pem -noout -text | grep "Issuer:" | awk '{print $7}' | cut -d '=' -f2 | sed 's/,$//g')

    ARRAY_OF_LOCALITIES=($(echo "$EXPECTED_ISSUER_LOCALITY" | tr ',' '\n'))

    for LOCALITY in ${ARRAY_OF_LOCALITIES[@]}
    do
      if [[ "$ISSUER_LOCALITY" == "$LOCALITY" ]]; then
        echo -e "${GREEN}Issuer locality of the client certificate is valid. Proceeding with the next steps... ${NC}"
        IS_CERT_ISSUER_LOCALITY_VALID=true
        break 2
      fi
    done

    echo -e "${RED}The issuer locality of the client certificate didn't match any of the expected ones. We expect one of \"$EXPECTED_ISSUER_LOCALITY\" but have: \"$ISSUER_LOCALITY\" ${NC}"
    echo -e "${YELLOW}[Retry $(($i+1))] Getting new client certificate... ${NC}"
    getClientCert # This will override the client certificate file content with the newly issued certificate
    sleep 0.2 # Sleep for 200ms before next retry
  done

  if [[ $IS_CERT_ISSUER_LOCALITY_VALID == false ]]; then
    echo -e "${RED}Couldn't get the client certificate with valid issuer locality after $CLIENT_CERT_RETRY_ATTEMPTS attempts. Exiting... ${NC}"
    exit 1
  fi
}

echo -e "${YELLOW}Issuing token... ${NC}"

if [[ "$EXT_NOT_EXTERNAL" == true ]]; then
  echo "$CERT_SVC_OAUTH_CLIENT_CERT" > /tmp/client-cert.pem
  echo "$CERT_SVC_OAUTH_CLIENT_KEY" > /tmp/client-key.pem
else
 echo "$CERT_SVC_OAUTH_CLIENT_CERT" | openssl enc -base64 -d -A -out /tmp/client-cert.pem
 echo "$CERT_SVC_OAUTH_CLIENT_KEY" | openssl enc -base64 -d -A -out /tmp/client-key.pem
fi

TOKEN=$(curl \
  -s $SKIP_SSL_VALIDATION_FLAG \
  -m 30 \
  -X POST \
  --cert /tmp/client-cert.pem \
  --key /tmp/client-key.pem \
  "$CERT_SVC_OAUTH_URL$CERT_SVC_TOKEN_PATH" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "Accept: application/json" \
  -d "grant_type=client_credentials&token_format=bearer&client_id=$CERT_SVC_CLIENT_ID" \
  | jq -r .access_token)

if [[ -z "$TOKEN" || $TOKEN == "null" ]]; then
  echo -e "${RED}Bearer token should not be empty or null. Exiting... ${NC}"
  exit 1
fi

echo -e "${YELLOW}Generating an encrypted private key... ${NC}"
PASS_PHRASE=$(openssl rand -base64 32)
openssl genpkey -pass pass:"$PASS_PHRASE" -aes-256-cbc -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out /tmp/encrypted-private-key.pem

echo -e "${YELLOW}Saving an unencrypted copy of the private key... ${NC}" # Later we use it to create a k8s secret, currently k8s does not support loading encrypted private keys
openssl rsa -in /tmp/encrypted-private-key.pem -out /tmp/unencrypted-private-key.pem -passin pass:"$PASS_PHRASE"

echo -e "${YELLOW}Creating a CSR with the following subject: $CERT_SUBJECT_PATTERN ${NC}"
openssl req -new -sha256 -key /tmp/encrypted-private-key.pem -passin pass:"$PASS_PHRASE" -out /tmp/my-csr.pem -subj "$CERT_SUBJECT_PATTERN"
JSON_CSR=$(jq -sR '.' /tmp/my-csr.pem)
echo -e "${YELLOW}Created CSR: $JSON_CSR ${NC}"

getClientCert

confirmValidIssuerLocalityOrRetry

echo -e "${YELLOW}Creating/Updating client certificate secret... ${NC}"
kubectl create secret generic "$CLIENT_CERT_SECRET_NAME" --namespace=compass-system --from-literal="$CLIENT_CERT_CERT_KEY"="$(cat /tmp/client-certificate_pkcs7.pem)" --from-literal="$CLIENT_CERT_KEY_KEY"="$(cat /tmp/unencrypted-private-key.pem)" --save-config --dry-run=client -o yaml | kubectl apply -f -

set +e