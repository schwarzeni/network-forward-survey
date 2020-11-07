go build .

ip_list=(10.211.55.54 10.211.55.57 10.211.55.58 10.211.55.59 10.211.55.60 10.211.55.61)
org_path=$(pwd)
for ip in "${ip_list[@]}";
do
  scp ws-server root@$ip:/root/
done

cd server-ssl
for ip in "${ip_list[@]}";
do
  mkdir server-ssl
  cp ca/* server/$ip/* server-ssl/
  scp -r server-ssl root@$ip:/root/
  rm -rf server-ssl
done

# for current dev server
cp ca/* server/10.211.55.52/* ./


cd $org_path



# scp ws-server root@10.211.55.54:/root/
# scp ws-server root@10.211.55.57:/root/
# scp ws-server root@10.211.55.58:/root/
# scp ws-server root@10.211.55.59:/root/
# scp ws-server root@10.211.55.60:/root/
# scp ws-server root@10.211.55.61:/root/

# scp -r server-ssl root@10.211.55.54:/root/
