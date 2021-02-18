rm -r analysis-api/infrastructure/proto/certs/*.pem
rm -r analysis-worker/proto/certs/*.pem

openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout analysis-api/infrastructure/proto/certs/ca-key.pem -out analysis-api/infrastructure/proto/certs/ca-cert.pem -subj "/C=BG/ST=Sofia/L=Sofia/O=dystopia.systems/OU=N\/A/CN=*.dystopia.systems/emailAddress=master@dystopia.systems"

echo "CA's self-signed certificate"
openssl x509 -in analysis-api/infrastructure/proto/certs/ca-cert.pem -noout -text

# 2. Generate web server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout analysis-worker/proto/certs/server-key.pem -out analysis-worker/proto/certs/server-req.pem -subj "/C=BG/ST=Sofia/L=Sofia/O=dystopia.systems/OU=N\/A/CN=*.dystopia.systems/emailAddress=master@dystopia.systems"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in analysis-worker/proto/certs/server-req.pem -days 60 -CA analysis-api/infrastructure/proto/certs/ca-cert.pem -CAkey analysis-api/infrastructure/proto/certs/ca-key.pem -CAcreateserial -out analysis-worker/proto/certs/server-cert.pem -extfile analysis-worker/proto/certs/server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in analysis-worker/proto/certs/server-cert.pem -noout -text
