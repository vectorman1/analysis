rm -r analysis-api/certs/*.pem
rm -r analysis-worker/certs/*.pem

mkdir -p analysis-api/certs
mkdir -p analysis-worker/certs

openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout analysis-api/certs/ca-key.pem -out analysis-api/certs/ca-cert.pem -subj "/C=BG/ST=Sofia/L=Sofia/O=dystopia.systems/OU=N\/A/CN=*.dystopia.systems/emailAddress=master@dystopia.systems"

echo "CA's self-signed certificate"
openssl x509 -in analysis-api/certs/ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout analysis-worker/certs/server-key.pem -out analysis-worker/certs/server-req.pem -subj "/C=BG/ST=Sofia/L=Sofia/O=dystopia.systems/OU=N\/A/CN=*.dystopia.systems/emailAddress=master@dystopia.systems"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in analysis-worker/certs/server-req.pem -days 60 -CA analysis-api/certs/ca-cert.pem -CAkey analysis-api/certs/ca-key.pem -CAcreateserial -out analysis-worker/certs/server-cert.pem -extfile analysis-worker/certs/server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in analysis-worker/certs/server-cert.pem -noout -text
