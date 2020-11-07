function ca_gen() {
  openssl genrsa -out ca/ca.key 2048
  openssl req -x509 -new -nodes -key ca/ca.key -days 10000 -out ca/ca.crt -subj "/CN=self-ca"
}

function server_gen() {
  server_ip=$1
  mkdir $server_ip
  cd $server_ip
  openssl genrsa -out server.key 2048
  openssl req -nodes -new -key server.key -subj "/CN=localhost"  -out server.csr
  echo subjectAltName = IP:$server_ip > extfile.cnf
  openssl x509 -req -in server.csr -CA ../../ca/ca.crt -CAkey ../../ca/ca.key -CAcreateserial -out server.crt -days 365 -extfile extfile.cnf
  cd ../
}

# ==== begin ====

org_path=$(pwd)
cd server-ssl

rm -rf *

mkdir ca
mkdir server

ca_gen

ip_list=(10.211.55.52 10.211.55.54 10.211.55.57 10.211.55.58 10.211.55.59 10.211.55.60 10.211.55.61)

cd server

for ip in "${ip_list[@]}"
do
  server_gen $ip
done

cd $org_path



