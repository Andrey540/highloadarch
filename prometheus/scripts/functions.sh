#!/usr/bin/env bash

source ./scripts/environmets.sh > /dev/null 2>&1 || source environmets.sh > /dev/null 2>&1

# Global functions
function Done() {
    echo -e '==> \E[32m'"\033\done\033[0m"
}

function Skip() {
    echo -e '==> \E[32m'"\033\skipped\033[0m"
}

function Failed() {
    echo -e '==> \E[91m'"\033\ failed\033[0m"
}

function EchoDash() {
echo "----------------------------------------------------------------"
}

function FinisMessage() {
        echo ""
        echo -e '\E[1m'"\033\Zabbix installation successfuly finished.\033[0m"
        echo "-----------------------------------------------------------------"
        echo ""
        echo -e '\E[1m'"\033\Zabbix UI is accessible at http://ip:8081 \033[0m"
        echo -e '\E[1m'"\033\Username: Admin \033[0m"
        echo -e '\E[1m'"\033\Pasword: zabbix (Don't forget to change it!)\033[0m"
        echo ""
        echo -e '\E[1m'"\033\Grafana UI is accessible at http://ip:3000 \033[0m"
        echo -e '\E[1m'"\033\Username: admin \033[0m"
        echo -e '\E[1m'"\033\Pasword: zabbix (Don't forget to change it too!)\033[0m"
        echo "-----------------------------------------------------------------"
}

### Grafana related
function CreateGRFAPIKey() {
    GRF_API_KEY=$(curl --insecure -s \
    -H "Accept: application/json" \
    -H "Content-Type:application/json" \
    -X POST -d \
     '{
	    "name":"zabbix-api-key",
	    "role": "Admin"
      }' \
     $GRF_SERVER_URL/api/auth/keys |jq .key |tr -d '"')
}

function CreatePrometheusDataSourcePD() {
cat <<EOF
{
        "orgId": 2,
        "name": "prometheus",
        "type": "prometheus",
        "access": "proxy",
        "url": "http://prometheus:9090",
        "jsonData": {
            "httpMethod": "POST"
        }
}
EOF
}

